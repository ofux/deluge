package worker

type Manager interface {
	CreateAll(jobShell *JobShell) error
	StartAll(jobShell *JobShell) error
}

type JobShell struct {
	ID       string
	DelugeID string
}

func GetManager() Manager {
	return &inMemoryManager{}
}
