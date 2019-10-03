package worker

import (
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/core/recording"
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

	go func() {
		for newStatus := range dlg.OnStatusChangeChan() {
			report := &repov2.PersistedWorkerReport{
				WorkerID:  m.globalWorkerID,
				JobID:     jobShell.ID,
				Status:    newStatus,
				Scenarios: make(map[string]*repov2.PersistedWorkerScenarioReport),
			}
			for scenarioID, scenario := range dlg.Scenarios {
				if scenario.Records == nil {
					continue
				}

				records, err := recording.MapHTTPRecords(scenario.Records)
				if err != nil {
					logrus.WithError(err).Errorf("Failed to map records of worker %s for scenario %s", m.globalWorkerID, scenarioID)
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
				logrus.WithError(err).Errorf("Failed to save records of worker %s", m.globalWorkerID)
			}
			//TODO: tell orchestrator that this worker is done
		}
	}()

	dlg.Run()

	return nil
}

func (m *inMemoryManager) InterruptAll(jobShellID string) error {
	m.runningDeluge.Interrupt()
	return nil
}
