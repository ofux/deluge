package worker

import (
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/repo"
	"github.com/ofux/deluge/repov2"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

type inMemoryManager struct {
	globalWorkerID string
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

	repo.Jobs.Store(&repo.RunningJob{
		ID:             jobShell.ID,
		RunnableDeluge: dlg,
	})

	go func() {
		<-dlg.Run()

		report := &repov2.PersistedWorkerReport{
			WorkerID: m.globalWorkerID,
			JobID:    jobShell.ID,
		}
		for scenarioID, scenario := range dlg.Scenarios {
			records, err := repov2.MapHTTPRecords(scenario.Records)
			if err != nil {
				logrus.WithError(err).Errorf("Failed to map records of scenario %s", scenarioID)
				return
			}
			report.Records[scenarioID] = records
		}
		err := repov2.Instance.SaveWorkerReport(report)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to save records of worker %s", m.globalWorkerID)
		}
	}()

	return nil
}
