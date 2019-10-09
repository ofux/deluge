package worker

import (
	uuid "github.com/satori/go.uuid"
)

type inMemoryManager struct {
	globalWorker *worker
}

func (m *inMemoryManager) CreateAll(jobShell *JobShell) error {
	m.globalWorker = NewWorker(uuid.NewV4().String(), jobShell)
	return nil
}

func (m *inMemoryManager) StartAll(jobShell *JobShell) error {
	return m.globalWorker.start()
}

func (m *inMemoryManager) InterruptAll(jobShellID string) error {
	m.globalWorker.interrupt()
	return nil
}
