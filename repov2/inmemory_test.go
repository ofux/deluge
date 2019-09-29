package repov2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScenarioDefinitionsRepository_Save(t *testing.T) {
	t.Run("Save one scenario with ID", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()
		const givenID = "givenID"
		scenario := &PersistedScenario{ID: givenID}

		err := testedRepo.SaveScenario(scenario)

		assert.NoError(t, err)
		assert.Len(t, testedRepo.scenarioDefinitions, 1)
		assert.Contains(t, testedRepo.scenarioDefinitions, givenID)
		assert.Equal(t, testedRepo.scenarioDefinitions[givenID], scenario)
	})

	t.Run("Save 3 scenarios with different IDs", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()
		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"
		scenario1 := &PersistedScenario{ID: givenID1}
		scenario2 := &PersistedScenario{ID: givenID2}
		scenario3 := &PersistedScenario{ID: givenID3}

		err := testedRepo.SaveScenario(scenario1)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.scenarioDefinitions, 1)
		assert.Contains(t, testedRepo.scenarioDefinitions, givenID1)
		assert.Equal(t, testedRepo.scenarioDefinitions[givenID1], scenario1)

		err = testedRepo.SaveScenario(scenario2)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.scenarioDefinitions, 2)
		assert.Contains(t, testedRepo.scenarioDefinitions, givenID2)
		assert.Equal(t, testedRepo.scenarioDefinitions[givenID2], scenario2)

		err = testedRepo.SaveScenario(scenario3)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.scenarioDefinitions, 3)
		assert.Contains(t, testedRepo.scenarioDefinitions, givenID3)
		assert.Equal(t, testedRepo.scenarioDefinitions[givenID3], scenario3)
	})

	t.Run("Save 2 scenarios with the same ID", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()
		const givenID = "givenID"
		scenario1 := &PersistedScenario{ID: givenID}
		scenario2 := &PersistedScenario{ID: givenID}

		err := testedRepo.SaveScenario(scenario1)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.scenarioDefinitions, 1)
		assert.Contains(t, testedRepo.scenarioDefinitions, givenID)
		assert.Equal(t, testedRepo.scenarioDefinitions[givenID], scenario1)

		err = testedRepo.SaveScenario(scenario1)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.scenarioDefinitions, 1)
		assert.Contains(t, testedRepo.scenarioDefinitions, givenID)
		assert.Equal(t, testedRepo.scenarioDefinitions[givenID], scenario2)
	})
}

func TestScenarioDefinitionsRepository_Get(t *testing.T) {
	t.Run("Create 3 scenarios and Get the second one", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()
		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"
		scenario1 := &PersistedScenario{ID: givenID1}
		scenario2 := &PersistedScenario{ID: givenID2}
		scenario3 := &PersistedScenario{ID: givenID3}
		err := testedRepo.SaveScenario(scenario1)
		assert.NoError(t, err)
		err = testedRepo.SaveScenario(scenario2)
		assert.NoError(t, err)
		err = testedRepo.SaveScenario(scenario3)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.scenarioDefinitions, 3)

		retrievedScenario, ok := testedRepo.GetScenario(givenID2)
		assert.True(t, ok)
		assert.Equal(t, scenario2, retrievedScenario)
	})

	t.Run("Get a job that does not exist", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()

		_, ok := testedRepo.GetScenario("doesNotExist")
		assert.False(t, ok)
	})
}

func TestScenarioDefinitionsRepository_GetAll(t *testing.T) {
	t.Run("Create 3 job and Get all of them", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()
		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"
		scenario1 := &PersistedScenario{ID: givenID1}
		scenario2 := &PersistedScenario{ID: givenID2}
		scenario3 := &PersistedScenario{ID: givenID3}
		err := testedRepo.SaveScenario(scenario1)
		assert.NoError(t, err)
		err = testedRepo.SaveScenario(scenario2)
		assert.NoError(t, err)
		err = testedRepo.SaveScenario(scenario3)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.scenarioDefinitions, 3)

		retrievedScenarios := testedRepo.GetAllScenarios()
		assert.Len(t, retrievedScenarios, 3)
		assert.Contains(t, retrievedScenarios, scenario1)
		assert.Contains(t, retrievedScenarios, scenario2)
		assert.Contains(t, retrievedScenarios, scenario3)
	})

	t.Run("Get all jobs of an empty repo", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()

		retrievedDlgs := testedRepo.GetAllScenarios()
		assert.NotNil(t, retrievedDlgs)
		assert.Len(t, retrievedDlgs, 0)
	})
}

func TestScenarioDefinitionsRepository_Delete(t *testing.T) {
	t.Run("Create 3 job and delete the second one", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()
		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"
		scenario1 := &PersistedScenario{ID: givenID1}
		scenario2 := &PersistedScenario{ID: givenID2}
		scenario3 := &PersistedScenario{ID: givenID3}
		err := testedRepo.SaveScenario(scenario1)
		assert.NoError(t, err)
		err = testedRepo.SaveScenario(scenario2)
		assert.NoError(t, err)
		err = testedRepo.SaveScenario(scenario3)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.scenarioDefinitions, 3)

		ok := testedRepo.DeleteScenario(givenID2)
		assert.True(t, ok)
		assert.Len(t, testedRepo.scenarioDefinitions, 2)
		assert.Contains(t, testedRepo.scenarioDefinitions, givenID1)
		assert.NotContains(t, testedRepo.scenarioDefinitions, givenID2)
		assert.Contains(t, testedRepo.scenarioDefinitions, givenID3)
	})

	t.Run("Delete a job that does not exist", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()

		ok := testedRepo.DeleteScenario("doesNotExist")
		assert.False(t, ok)
	})
}

// TestScenarioDefinitionsRepository_Race is only meant to be run with the go race detector tool
func TestScenarioDefinitionsRepository_Race(t *testing.T) {
	testedRepo := NewInMemoryRepository()
	const givenID1 = "givenID1"
	const givenID2 = "givenID2"
	scenario1 := &PersistedScenario{ID: givenID1}
	scenario2 := &PersistedScenario{ID: givenID2}

	go func() {
		err := testedRepo.SaveScenario(scenario1)
		assert.NoError(t, err)
		assert.Contains(t, testedRepo.scenarioDefinitions, givenID1)
	}()

	go func() {
		err := testedRepo.SaveScenario(scenario2)
		assert.NoError(t, err)
		assert.Contains(t, testedRepo.scenarioDefinitions, givenID2)
	}()

	go func() {
		testedRepo.DeleteScenario(givenID1)
	}()

	go func() {
		testedRepo.GetAllScenarios()
	}()

	go func() {
		testedRepo.GetScenario(givenID1)
	}()
}