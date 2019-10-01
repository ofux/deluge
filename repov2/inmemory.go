package repov2

import (
	"sync"
)

type InMemoryRepository struct {
	delugeDefinitions map[string]*PersistedDeluge
	mutDeluges        *sync.Mutex

	scenarioDefinitions map[string]*PersistedScenario
	mutScenarios        *sync.Mutex

	jobShells    map[string]*PersistedJobShell
	mutJobShells *sync.Mutex

	workerReports    map[string]*PersistedWorkerReport
	mutWorkerReports *sync.Mutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		delugeDefinitions:   make(map[string]*PersistedDeluge),
		mutDeluges:          &sync.Mutex{},
		scenarioDefinitions: make(map[string]*PersistedScenario),
		mutScenarios:        &sync.Mutex{},
		jobShells:           make(map[string]*PersistedJobShell),
		mutJobShells:        &sync.Mutex{},
		workerReports:       make(map[string]*PersistedWorkerReport),
		mutWorkerReports:    &sync.Mutex{},
	}
}

func (r *InMemoryRepository) SaveDeluge(deluge *PersistedDeluge) error {
	r.mutDeluges.Lock()
	defer r.mutDeluges.Unlock()
	r.delugeDefinitions[deluge.ID] = deluge
	return nil
}

func (r *InMemoryRepository) GetDeluge(id string) (*PersistedDeluge, bool) {
	r.mutDeluges.Lock()
	defer r.mutDeluges.Unlock()
	def, ok := r.delugeDefinitions[id]
	return def, ok
}

func (r *InMemoryRepository) GetAllDeluges() []*PersistedDeluge {
	r.mutDeluges.Lock()
	defer r.mutDeluges.Unlock()
	all := make([]*PersistedDeluge, 0, len(r.delugeDefinitions))
	for _, v := range r.delugeDefinitions {
		all = append(all, v)
	}
	return all
}

func (r *InMemoryRepository) DeleteDeluge(id string) bool {
	r.mutDeluges.Lock()
	defer r.mutDeluges.Unlock()
	if _, ok := r.delugeDefinitions[id]; ok {
		delete(r.delugeDefinitions, id)
		return true
	}
	return false
}

// ======

func (r *InMemoryRepository) SaveScenario(scenario *PersistedScenario) error {
	r.mutScenarios.Lock()
	defer r.mutScenarios.Unlock()
	r.scenarioDefinitions[scenario.ID] = scenario
	return nil
}

func (r *InMemoryRepository) GetScenario(id string) (*PersistedScenario, bool) {
	r.mutScenarios.Lock()
	defer r.mutScenarios.Unlock()
	def, ok := r.scenarioDefinitions[id]
	return def, ok
}

func (r *InMemoryRepository) GetDelugeScenarios(ids []string) map[string]*PersistedScenario {
	r.mutScenarios.Lock()
	defer r.mutScenarios.Unlock()
	delugeScenarios := make(map[string]*PersistedScenario)
	for _, id := range ids {
		if scenario, ok := r.scenarioDefinitions[id]; ok {
			delugeScenarios[id] = scenario
		}
	}
	return delugeScenarios
}

func (r *InMemoryRepository) GetAllScenarios() []*PersistedScenario {
	r.mutScenarios.Lock()
	defer r.mutScenarios.Unlock()
	all := make([]*PersistedScenario, 0, len(r.scenarioDefinitions))
	for _, v := range r.scenarioDefinitions {
		all = append(all, v)
	}
	return all
}

func (r *InMemoryRepository) DeleteScenario(id string) bool {
	r.mutScenarios.Lock()
	defer r.mutScenarios.Unlock()
	if _, ok := r.scenarioDefinitions[id]; ok {
		delete(r.scenarioDefinitions, id)
		return true
	}
	return false
}

// =======

func (r *InMemoryRepository) SaveJobShell(jobShell *PersistedJobShell) error {
	r.mutJobShells.Lock()
	defer r.mutJobShells.Unlock()
	r.jobShells[jobShell.ID] = jobShell
	return nil
}

func (r *InMemoryRepository) GetJobShell(id string) (*PersistedJobShell, bool) {
	r.mutJobShells.Lock()
	defer r.mutJobShells.Unlock()
	jobShell, ok := r.jobShells[id]
	return jobShell, ok
}

func (r *InMemoryRepository) GetAllJobShell() []*PersistedJobShell {
	r.mutJobShells.Lock()
	defer r.mutJobShells.Unlock()
	all := make([]*PersistedJobShell, 0, len(r.jobShells))
	for _, v := range r.jobShells {
		all = append(all, v)
	}
	return all
}

// WorkerReports

func (r *InMemoryRepository) SaveWorkerReport(workerReport *PersistedWorkerReport) error {
	r.mutWorkerReports.Lock()
	defer r.mutWorkerReports.Unlock()
	r.workerReports[workerReport.WorkerID] = workerReport
	return nil
}

func (r *InMemoryRepository) GetJobWorkerReports(jobID string) []*PersistedWorkerReport {
	r.mutWorkerReports.Lock()
	defer r.mutWorkerReports.Unlock()
	var reports []*PersistedWorkerReport
	for _, v := range r.workerReports {
		if v.JobID == jobID {
			reports = append(reports, v)
		}
	}
	return reports
}
