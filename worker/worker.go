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

var managerInstance Manager = &inMemoryManager{
	workers: make([]*worker, 1),
}

func GetManager() Manager {
	return managerInstance
}

type worker struct {
	ID            string
	jobShell      *JobShell
	runningDeluge *core.RunnableDeluge
}

func newWorker(ID string, jobShell *JobShell) *worker {
	return &worker{
		ID:       ID,
		jobShell: jobShell,
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
		report := &repov2.PersistedWorkerReport{
			WorkerID:  w.ID,
			JobID:     w.jobShell.ID,
			Status:    newStatus,
			Scenarios: make(map[string]*repov2.PersistedWorkerScenarioReport),
		}
		for scenarioID, scenario := range w.runningDeluge.Scenarios {
			if scenario.Records == nil {
				continue
			}

			records, err := recording.MapHTTPRecords(scenario.Records)
			if err != nil {
				logrus.WithError(err).Errorf("Failed to map records of worker %s for scenario %s", w.ID, scenarioID)
				continue
			}

			report.Scenarios[scenarioID] = &repov2.PersistedWorkerScenarioReport{
				Status:            scenario.Status,
				Errors:            scenario.Errors,
				IterationDuration: scenario.IterationDuration,
				Records:           records,
			}
		}
		err := repov2.Instance.SaveWorkerReport(report)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to save records of worker %s", w.ID)
		}
		//TODO: tell orchestrator that this worker is done
	}
}

func (w *worker) reportRecordsRegularly() {
	ticker := time.NewTicker(20 * time.Second)

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
			if scenarioSnapshot.Err != nil {
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

			records, err := recording.MapHTTPRecords(records)
			if err != nil {
				logrus.WithError(err).Errorf("Failed to map records of worker %s for scenario %s", w.ID, scenarioID)
				continue
			}

			report.Scenarios[scenarioID] = &repov2.PersistedWorkerScenarioReport{
				Records: records,
			}
		}
		err = repov2.Instance.SaveWorkerReport(report)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to save records of worker %s", w.ID)
		}
	}
}
