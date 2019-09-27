package repov2

import (
	"sync"
)

type PersistedDeluge struct {
	ID     string
	Name   string
	Script string
}

type PersistedScenario struct {
	ID     string
	Name   string
	Script string
}

type PersistedJobShell struct {
	ID       string
	DelugeID string
	Webhook  string
}

type Repository struct {
	delugeDefinitions map[string]*PersistedDeluge
	mutDeluges        *sync.Mutex

	scenarioDefinitions map[string]*PersistedScenario
	mutScenarios        *sync.Mutex

	jobShells    map[string]*PersistedJobShell
	mutJobShells *sync.Mutex
}

var Instance = NewInMemoryRepository()

func NewInMemoryRepository() *Repository {
	return &Repository{
		delugeDefinitions:   make(map[string]*PersistedDeluge),
		mutDeluges:          &sync.Mutex{},
		scenarioDefinitions: make(map[string]*PersistedScenario),
		mutScenarios:        &sync.Mutex{},
		jobShells:           make(map[string]*PersistedJobShell),
		mutJobShells:        &sync.Mutex{},
	}
}

func (jr *Repository) SaveDeluge(deluge *PersistedDeluge) error {
	jr.mutDeluges.Lock()
	defer jr.mutDeluges.Unlock()
	jr.delugeDefinitions[deluge.ID] = deluge
	return nil
}

func (jr *Repository) GetDeluge(id string) (*PersistedDeluge, bool) {
	jr.mutDeluges.Lock()
	defer jr.mutDeluges.Unlock()
	def, ok := jr.delugeDefinitions[id]
	return def, ok
}

func (jr *Repository) GetAllDeluges() []*PersistedDeluge {
	jr.mutDeluges.Lock()
	defer jr.mutDeluges.Unlock()
	all := make([]*PersistedDeluge, 0, len(jr.delugeDefinitions))
	for _, v := range jr.delugeDefinitions {
		all = append(all, v)
	}
	return all
}

func (jr *Repository) DeleteDeluge(id string) bool {
	jr.mutDeluges.Lock()
	defer jr.mutDeluges.Unlock()
	if _, ok := jr.delugeDefinitions[id]; ok {
		delete(jr.delugeDefinitions, id)
		return true
	}
	return false
}

// ======

func (jr *Repository) SaveScenario(scenario *PersistedScenario) error {
	jr.mutScenarios.Lock()
	defer jr.mutScenarios.Unlock()
	jr.scenarioDefinitions[scenario.ID] = scenario
	return nil
}

func (jr *Repository) GetScenario(id string) (*PersistedScenario, bool) {
	jr.mutScenarios.Lock()
	defer jr.mutScenarios.Unlock()
	def, ok := jr.scenarioDefinitions[id]
	return def, ok
}

func (jr *Repository) GetAllScenarios() []*PersistedScenario {
	jr.mutScenarios.Lock()
	defer jr.mutScenarios.Unlock()
	all := make([]*PersistedScenario, 0, len(jr.scenarioDefinitions))
	for _, v := range jr.scenarioDefinitions {
		all = append(all, v)
	}
	return all
}

func (jr *Repository) DeleteScenario(id string) bool {
	jr.mutScenarios.Lock()
	defer jr.mutScenarios.Unlock()
	if _, ok := jr.scenarioDefinitions[id]; ok {
		delete(jr.scenarioDefinitions, id)
		return true
	}
	return false
}

// =======

func (jr *Repository) SaveJobShell(jobShell *PersistedJobShell) error {
	jr.mutJobShells.Lock()
	defer jr.mutJobShells.Unlock()
	jr.jobShells[jobShell.ID] = jobShell
	return nil
}

func (jr *Repository) GetJobShell(id string) (*PersistedJobShell, bool) {
	jr.mutJobShells.Lock()
	defer jr.mutJobShells.Unlock()
	jobShell, ok := jr.jobShells[id]
	return jobShell, ok
}
