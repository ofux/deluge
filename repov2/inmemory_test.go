package repov2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// DELUGE

func TestInMemoryRepository_SaveDeluge(t *testing.T) {
	t.Run("Save one deluge with ID", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()
		const givenID = "givenID"
		deluge := &PersistedDeluge{ID: givenID}

		err := testedRepo.SaveDeluge(deluge)

		assert.NoError(t, err)
		assert.Len(t, testedRepo.delugeDefinitions, 1)
		assert.Contains(t, testedRepo.delugeDefinitions, givenID)
		assert.Equal(t, testedRepo.delugeDefinitions[givenID], deluge)
	})

	t.Run("Save 3 deluges with different IDs", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()
		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"
		deluge1 := &PersistedDeluge{ID: givenID1}
		deluge2 := &PersistedDeluge{ID: givenID2}
		deluge3 := &PersistedDeluge{ID: givenID3}

		err := testedRepo.SaveDeluge(deluge1)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.delugeDefinitions, 1)
		assert.Contains(t, testedRepo.delugeDefinitions, givenID1)
		assert.Equal(t, testedRepo.delugeDefinitions[givenID1], deluge1)

		err = testedRepo.SaveDeluge(deluge2)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.delugeDefinitions, 2)
		assert.Contains(t, testedRepo.delugeDefinitions, givenID2)
		assert.Equal(t, testedRepo.delugeDefinitions[givenID2], deluge2)

		err = testedRepo.SaveDeluge(deluge3)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.delugeDefinitions, 3)
		assert.Contains(t, testedRepo.delugeDefinitions, givenID3)
		assert.Equal(t, testedRepo.delugeDefinitions[givenID3], deluge3)
	})

	t.Run("Save 2 deluges with the same ID", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()
		const givenID = "givenID"
		deluge1 := &PersistedDeluge{ID: givenID}
		deluge2 := &PersistedDeluge{ID: givenID}

		err := testedRepo.SaveDeluge(deluge1)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.delugeDefinitions, 1)
		assert.Contains(t, testedRepo.delugeDefinitions, givenID)
		assert.Equal(t, testedRepo.delugeDefinitions[givenID], deluge1)

		err = testedRepo.SaveDeluge(deluge1)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.delugeDefinitions, 1)
		assert.Contains(t, testedRepo.delugeDefinitions, givenID)
		assert.Equal(t, testedRepo.delugeDefinitions[givenID], deluge2)
	})
}

func TestInMemoryRepository_GetDeluge(t *testing.T) {
	t.Run("Create 3 deluges and Get the second one", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()
		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"
		deluge1 := &PersistedDeluge{ID: givenID1}
		deluge2 := &PersistedDeluge{ID: givenID2}
		deluge3 := &PersistedDeluge{ID: givenID3}
		err := testedRepo.SaveDeluge(deluge1)
		assert.NoError(t, err)
		err = testedRepo.SaveDeluge(deluge2)
		assert.NoError(t, err)
		err = testedRepo.SaveDeluge(deluge3)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.delugeDefinitions, 3)

		retrievedDeluge, ok := testedRepo.GetDeluge(givenID2)
		assert.True(t, ok)
		assert.Equal(t, deluge2, retrievedDeluge)
	})

	t.Run("Get a deluge that does not exist", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()

		_, ok := testedRepo.GetDeluge("doesNotExist")
		assert.False(t, ok)
	})
}

func TestInMemoryRepository_GetAllDeluges(t *testing.T) {
	t.Run("Create 3 deluges and Get all of them", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()
		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"
		deluge1 := &PersistedDeluge{ID: givenID1}
		deluge2 := &PersistedDeluge{ID: givenID2}
		deluge3 := &PersistedDeluge{ID: givenID3}
		err := testedRepo.SaveDeluge(deluge1)
		assert.NoError(t, err)
		err = testedRepo.SaveDeluge(deluge2)
		assert.NoError(t, err)
		err = testedRepo.SaveDeluge(deluge3)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.delugeDefinitions, 3)

		retrievedDeluges := testedRepo.GetAllDeluges()
		assert.Len(t, retrievedDeluges, 3)
		assert.Contains(t, retrievedDeluges, deluge1)
		assert.Contains(t, retrievedDeluges, deluge2)
		assert.Contains(t, retrievedDeluges, deluge3)
	})

	t.Run("Get all deluges of an empty repo", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()

		retrievedDlgs := testedRepo.GetAllDeluges()
		assert.NotNil(t, retrievedDlgs)
		assert.Len(t, retrievedDlgs, 0)
	})
}

func TestInMemoryRepository_DeleteDeluge(t *testing.T) {
	t.Run("Create 3 deluges and delete the second one", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()
		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"
		deluge1 := &PersistedDeluge{ID: givenID1}
		deluge2 := &PersistedDeluge{ID: givenID2}
		deluge3 := &PersistedDeluge{ID: givenID3}
		err := testedRepo.SaveDeluge(deluge1)
		assert.NoError(t, err)
		err = testedRepo.SaveDeluge(deluge2)
		assert.NoError(t, err)
		err = testedRepo.SaveDeluge(deluge3)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.delugeDefinitions, 3)

		ok := testedRepo.DeleteDeluge(givenID2)
		assert.True(t, ok)
		assert.Len(t, testedRepo.delugeDefinitions, 2)
		assert.Contains(t, testedRepo.delugeDefinitions, givenID1)
		assert.NotContains(t, testedRepo.delugeDefinitions, givenID2)
		assert.Contains(t, testedRepo.delugeDefinitions, givenID3)
	})

	t.Run("Delete a deluge that does not exist", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()

		ok := testedRepo.DeleteDeluge("doesNotExist")
		assert.False(t, ok)
	})
}

// TestInMemoryRepository_Race is only meant to be run with the go race detector tool
func TestInMemoryRepository_RaceDeluge(t *testing.T) {
	testedRepo := NewInMemoryRepository()
	const givenID1 = "givenID1"
	const givenID2 = "givenID2"
	deluge1 := &PersistedDeluge{ID: givenID1}
	deluge2 := &PersistedDeluge{ID: givenID2}

	go func() {
		err := testedRepo.SaveDeluge(deluge1)
		assert.NoError(t, err)
		assert.Equal(t, testedRepo.delugeDefinitions[givenID1], deluge1)
	}()

	go func() {
		err := testedRepo.SaveDeluge(deluge2)
		assert.NoError(t, err)
		assert.Equal(t, testedRepo.delugeDefinitions[givenID2], deluge2)
	}()

	go func() {
		testedRepo.DeleteDeluge(givenID1)
	}()

	go func() {
		testedRepo.GetAllDeluges()
	}()

	go func() {
		testedRepo.GetDeluge(givenID1)
	}()
}

// SCENARIO

func TestInMemoryRepository_SaveScenario(t *testing.T) {
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

func TestInMemoryRepository_GetScenario(t *testing.T) {
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

func TestInMemoryRepository_GetDelugeScenarios(t *testing.T) {
	t.Run("Create 3 scenarios and Get the first and second one", func(t *testing.T) {
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

		retrievedScenarios := testedRepo.GetDelugeScenarios([]string{givenID1, givenID2})
		assert.Len(t, retrievedScenarios, 2)
		assert.Equal(t, scenario1, retrievedScenarios[givenID1])
		assert.Equal(t, scenario2, retrievedScenarios[givenID2])
	})

	t.Run("Get a job that does not exist", func(t *testing.T) {
		testedRepo := NewInMemoryRepository()

		retrievedScenarios := testedRepo.GetDelugeScenarios([]string{"doesNotExist"})
		assert.Len(t, retrievedScenarios, 0)
	})
}

func TestInMemoryRepository_GetAllScenarios(t *testing.T) {
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

func TestInMemoryRepository_DeleteScenario(t *testing.T) {
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

// TestInMemoryRepository_Race is only meant to be run with the go race detector tool
func TestInMemoryRepository_RaceScenario(t *testing.T) {
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
		testedRepo.GetDelugeScenarios([]string{givenID1, givenID2})
	}()

	go func() {
		testedRepo.GetScenario(givenID1)
	}()
}
