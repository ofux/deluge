package worker

import (
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/repov2"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

type inMemoryManager struct {
	globalWorkerID string
	runningDeluge  *core.RunnableDeluge
}

func (m *inMemoryManager) CreateAll(jobShell *JobShell) error {
	m.globalWorkerID = uuid.NewV4().String()
	return nil
}

func (m *inMemoryManager) StartAll(jobShell *JobShell) error {
	return m.start(jobShell)
}

func (m *inMemoryManager) start(jobShell *JobShell) error {
	dlg, err := core.NewRunnableDeluge(jobShell.DelugeID)
	if err != nil {
		return errors.Wrapf(err, "failed to create runnable deluge from jobShell %s (delugeId %s)", jobShell.ID, jobShell.DelugeID)
	}
	m.runningDeluge = dlg

	// Store empty worker report
	reportShell := &repov2.PersistedWorkerReport{
		WorkerID: m.globalWorkerID,
		JobID:    jobShell.ID,
		Status:   dlg.Status,
	}
	err = repov2.Instance.SaveWorkerReport(reportShell)
	if err != nil {
		return errors.Wrapf(err, "failed to create worker report '%s' for jobShell %s (delugeId %s)", reportShell.WorkerID, reportShell.JobID, jobShell.DelugeID)
	}

	go func() {
		for range dlg.OnStatusChangeChan() {
			report := &repov2.PersistedWorkerReport{
				WorkerID:  m.globalWorkerID,
				JobID:     jobShell.ID,
				Status:    dlg.Status,
				Scenarios: make(map[string]*repov2.PersistedWorkerScenarioReport),
			}
			for scenarioID, scenario := range dlg.Scenarios {
				if scenario.Records == nil {
					continue
				}
				records, err := repov2.MapHTTPRecords(scenario.Records)
				if err != nil {
					logrus.WithError(err).Errorf("Failed to map records of scenario %s", scenarioID)
					return
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
				logrus.WithError(err).Errorf("Failed to save records of worker %s", m.globalWorkerID)
			}
		}
	}()

	dlg.Run()

	return nil
}

func (m *inMemoryManager) InterruptAll(jobShellID string) error {
	m.runningDeluge.Interrupt()
	return nil
}
