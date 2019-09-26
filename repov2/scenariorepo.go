package repov2

import (
	"sync"
)

type PersistedScenario struct {
	ID     string
	Name   string
	Script string
}

type ScenarioDefinitionsRepository struct {
	scenarioDefinitions map[string]*PersistedScenario
	mut                 *sync.Mutex
}

var ScenarioDefinitions = NewScenarioDefinitionsRepository()

func NewScenarioDefinitionsRepository() *ScenarioDefinitionsRepository {
	return &ScenarioDefinitionsRepository{
		scenarioDefinitions: make(map[string]*PersistedScenario),
		mut:                 &sync.Mutex{},
	}
}

func (jr *ScenarioDefinitionsRepository) Save(scenario *PersistedScenario) error {
	jr.mut.Lock()
	defer jr.mut.Unlock()
	jr.scenarioDefinitions[scenario.ID] = scenario
	return nil
}

func (jr *ScenarioDefinitionsRepository) Get(id string) (*PersistedScenario, bool) {
	jr.mut.Lock()
	defer jr.mut.Unlock()
	def, ok := jr.scenarioDefinitions[id]
	return def, ok
}

func (jr *ScenarioDefinitionsRepository) GetAll() []*PersistedScenario {
	jr.mut.Lock()
	defer jr.mut.Unlock()
	all := make([]*PersistedScenario, 0, len(jr.scenarioDefinitions))
	for _, v := range jr.scenarioDefinitions {
		all = append(all, v)
	}
	return all
}

func (jr *ScenarioDefinitionsRepository) Delete(id string) bool {
	jr.mut.Lock()
	defer jr.mut.Unlock()
	if _, ok := jr.scenarioDefinitions[id]; ok {
		delete(jr.scenarioDefinitions, id)
		return true
	}
	return false
}
