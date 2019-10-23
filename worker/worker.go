package worker

import (
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/repov2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type Manager interface {
	CreateAll(jobShell *JobShell) error
	StartAll(jobShell *JobShell) error
	InterruptAll(jobShellID string) error
}

type JobShell struct {
	ID       string
	DelugeID string
}

var ManagerInstance Manager = NewInMemoryManager(1)

func GetManager() Manager {
	return ManagerInstance
}

type worker struct {
	ID            string
	jobShell      *JobShell
	runningDeluge *core.RunnableDeluge
	repository    repov2.Repository

	regularReportFrequency time.Duration
	finalReportRetryCount  int
	finalReportRetryDelay  time.Duration

	logger *logrus.Entry
}

func newWorker(ID string, jobShell *JobShell, repository repov2.Repository) *worker {
	return &worker{
		ID:         ID,
		jobShell:   jobShell,
		repository: repository,

		regularReportFrequency: 20 * time.Second,
		finalReportRetryCount:  3,
		finalReportRetryDelay:  10 * time.Second,

		logger: logrus.WithField("workerId", ID).WithField("jobId", jobShell.ID).WithField("delugeId", jobShell.DelugeID),
	}
}

func (w *worker) interrupt() {
	w.runningDeluge.Interrupt()
}

func (w *worker) start() error {
	dlg, err := core.NewRunnableDeluge(w.jobShell.DelugeID)
	if err != nil {
		return errors.Wrapf(err, "failed to create runnable deluge from jobShell %s (delugeId %s)", w.jobShell.ID, w.jobShell.DelugeID)
	}
	w.runningDeluge = dlg

	go func() {
		w.listenToStatusChanges()
	}()

	go func() {
		w.reportRecordsRegularly()
	}()

	dlg.Run()

	return nil
}

func (w *worker) listenToStatusChanges() {
	for newStatus := range w.runningDeluge.OnStatusChangeChan() {
		w.logger.Debugf("Received status %s", newStatus)

		report := &repov2.PersistedWorkerReport{
			WorkerID:  w.ID,
			JobID:     w.jobShell.ID,
			Status:    newStatus,
			Scenarios: make(map[string]*repov2.PersistedWorkerScenarioReport),
		}
		if newStatus.IsEnd() {
			for scenarioID, scenario := range w.runningDeluge.Scenarios {
				logger := w.logger.WithField("scenarioId", scenarioID)
				logger.Debug("Adding records of scenario")

				if scenario.Records == nil {
					logger.Debug("Scenario has no record")
					continue
				}

				records, err := recording.MapHTTPRecords(scenario.Records)
				if err != nil {
					logger.WithError(err).Error("Failed to map records of worker for scenario")
					continue
				}

				report.Scenarios[scenarioID] = &repov2.PersistedWorkerScenarioReport{
					Status:            scenario.Status,
					Errors:            scenario.Errors,
					IterationDuration: scenario.IterationDuration,
					Records:           records,
				}
				logger.Debug("Added records of scenario")
			}
			w.saveWorkerReportWithRetry(report)
		} else {
			w.saveWorkerReport(report)
		}
		//TODO: tell orchestrator that this worker is done
	}
}

func (w *worker) saveWorkerReport(report *repov2.PersistedWorkerReport) {
	err := w.repository.SaveWorkerReport(report)
	if err != nil {
		w.logger.WithError(err).Error("Failed to save records of worker")
	} else {
		w.logger.Debug("Successfully saved records of worker")
	}
}

func (w *worker) saveWorkerReportWithRetry(report *repov2.PersistedWorkerReport) {
	err := w.repository.SaveWorkerReport(report)

	const retryDelayMultiply = 3

	delay := w.finalReportRetryDelay
	for retry := 1; err != nil && retry <= w.finalReportRetryCount; retry++ {
		w.logger.WithError(err).Errorf("Failed to save records of worker. Retrying in %s", delay)
		time.Sleep(delay)
		err = w.repository.SaveWorkerReport(report)
		delay *= retryDelayMultiply
	}

	if err != nil {
		w.logger.WithError(err).Error("Failed to save records of worker")
	} else {
		w.logger.Debug("Successfully saved records of worker")
	}
}

func (w *worker) reportRecordsRegularly() {
	ticker := time.NewTicker(w.regularReportFrequency)

	allRecords := make(map[string]*recording.HTTPRecordsOverTime)

	for range ticker.C {
		snapshot, err := w.runningDeluge.GetRecordsSnapshot()
		if err != nil {
			ticker.Stop()
			break
		}
		for scenarioID, scenarioSnapshot := range snapshot {
			scenarioRecords, ok := allRecords[scenarioID]
			if !ok {
				scenarioRecords = &recording.HTTPRecordsOverTime{}
				allRecords[scenarioID] = scenarioRecords
			}
			if scenarioSnapshot.Err != nil || scenarioSnapshot.HTTPRecordsOverTimeSnapshot == nil {
				continue
			}
			scenarioRecords.Global = scenarioSnapshot.HTTPRecordsOverTimeSnapshot.Global
			for overTimeIndex, rec := range scenarioSnapshot.HTTPRecordsOverTimeSnapshot.OverTime {
				if len(scenarioRecords.OverTime) <= overTimeIndex {
					scenarioRecords.OverTime = append(scenarioRecords.OverTime, make([]*recording.HTTPRecord, overTimeIndex+1-len(scenarioRecords.OverTime))...)
				}
				scenarioRecords.OverTime[overTimeIndex] = rec
			}
		}

		report := &repov2.PersistedWorkerReport{
			WorkerID:  w.ID,
			JobID:     w.jobShell.ID,
			Scenarios: make(map[string]*repov2.PersistedWorkerScenarioReport),
		}
		for scenarioID, records := range allRecords {
			if records == nil {
				continue
			}

			logger := w.logger.WithField("scenarioId", scenarioID)

			records, err := recording.MapHTTPRecords(records)
			if err != nil {
				logger.WithError(err).Error("Failed to map records of worker for scenario")
				continue
			}

			report.Scenarios[scenarioID] = &repov2.PersistedWorkerScenarioReport{
				Records: records,
			}
		}
		w.saveWorkerReport(report)
	}
}
