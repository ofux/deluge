package worker

import (
	"github.com/ofux/deluge/repov2"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type inMemoryManager struct {
	workers []*worker
}

func NewInMemoryManager(workerCount int) Manager {
	return &inMemoryManager{
		workers: make([]*worker, workerCount),
	}
}

func (m *inMemoryManager) CreateAll(jobShell *JobShell) error {
	for i := range m.workers {
		m.workers[i] = newWorker(uuid.NewV4().String(), jobShell, repov2.Instance)
	}
	return nil
}

func (m *inMemoryManager) StartAll(jobShell *JobShell) error {
	for _, w := range m.workers {
		if err := w.start(); err != nil {
			return errors.Wrapf(err, "failed to start worker %s", w.ID)
		}
	}
	return nil
}

func (m *inMemoryManager) InterruptAll(jobShellID string) error {
	for _, w := range m.workers {
		w.interrupt()
	}
	return nil
}
