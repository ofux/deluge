package repo

import (
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/parser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJobsRepository_CreateWithID(t *testing.T) {
	t.Run("Create one job with ID", func(t *testing.T) {
		prg := getTestProgram(t)
		testedRepo := NewJobsRepository()
		const givenID = "givenID"

		dlg, err := testedRepo.CreateWithID(prg, givenID)

		assert.NoError(t, err)
		assert.NotNil(t, dlg)
		assert.Len(t, testedRepo.jobs, 1)
		assert.Contains(t, testedRepo.jobs, givenID)
		assert.Equal(t, testedRepo.jobs[givenID], dlg)
		assert.Equal(t, dlg.ID, givenID)
	})

	t.Run("Create 3 job with different IDs", func(t *testing.T) {
		prg := getTestProgram(t)
		testedRepo := NewJobsRepository()
		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"

		dlg, err := testedRepo.CreateWithID(prg, givenID1)
		assert.NoError(t, err)
		assert.NotNil(t, dlg)
		assert.Len(t, testedRepo.jobs, 1)
		assert.Contains(t, testedRepo.jobs, givenID1)
		assert.Equal(t, testedRepo.jobs[givenID1], dlg)
		assert.Equal(t, dlg.ID, givenID1)

		dlg, err = testedRepo.CreateWithID(prg, givenID2)
		assert.NoError(t, err)
		assert.NotNil(t, dlg)
		assert.Len(t, testedRepo.jobs, 2)
		assert.Contains(t, testedRepo.jobs, givenID2)
		assert.Equal(t, testedRepo.jobs[givenID2], dlg)
		assert.Equal(t, dlg.ID, givenID2)

		dlg, err = testedRepo.CreateWithID(prg, givenID3)
		assert.NoError(t, err)
		assert.NotNil(t, dlg)
		assert.Len(t, testedRepo.jobs, 3)
		assert.Contains(t, testedRepo.jobs, givenID3)
		assert.Equal(t, testedRepo.jobs[givenID3], dlg)
		assert.Equal(t, dlg.ID, givenID3)
	})

	t.Run("Create 2 jobs with the same ID", func(t *testing.T) {
		prg := getTestProgram(t)
		testedRepo := NewJobsRepository()
		const givenID = "givenID"

		dlg, err := testedRepo.CreateWithID(prg, givenID)

		assert.NoError(t, err)
		assert.NotNil(t, dlg)
		assert.Len(t, testedRepo.jobs, 1)
		assert.Contains(t, testedRepo.jobs, givenID)
		assert.Equal(t, testedRepo.jobs[givenID], dlg)
		assert.Equal(t, dlg.ID, givenID)

		_, err = testedRepo.CreateWithID(prg, givenID)

		assert.Error(t, err)
		assert.Len(t, testedRepo.jobs, 1)
		assert.Contains(t, testedRepo.jobs, givenID)
		assert.Equal(t, testedRepo.jobs[givenID], dlg)
		assert.Equal(t, dlg.ID, givenID)
	})

	t.Run("Create one job with an error", func(t *testing.T) {
		prg := getTestProgramWithError(t)
		testedRepo := NewJobsRepository()
		const givenID = "givenID"

		_, err := testedRepo.CreateWithID(prg, givenID)

		assert.Error(t, err)
		assert.Len(t, testedRepo.jobs, 0)
	})
}

func TestJobsRepository_Create(t *testing.T) {
	t.Run("Create 3 job with generated IDs", func(t *testing.T) {
		prg := getTestProgram(t)
		testedRepo := NewJobsRepository()

		dlg, err := testedRepo.Create(prg)
		assert.NoError(t, err)
		assert.NotNil(t, dlg)
		assert.Len(t, testedRepo.jobs, 1)
		assert.Contains(t, testedRepo.jobs, dlg.ID)
		assert.Equal(t, testedRepo.jobs[dlg.ID], dlg)

		dlg, err = testedRepo.Create(prg)
		assert.NoError(t, err)
		assert.NotNil(t, dlg)
		assert.Len(t, testedRepo.jobs, 2)
		assert.Contains(t, testedRepo.jobs, dlg.ID)
		assert.Equal(t, testedRepo.jobs[dlg.ID], dlg)

		dlg, err = testedRepo.Create(prg)
		assert.NoError(t, err)
		assert.NotNil(t, dlg)
		assert.Len(t, testedRepo.jobs, 3)
		assert.Contains(t, testedRepo.jobs, dlg.ID)
		assert.Equal(t, testedRepo.jobs[dlg.ID], dlg)
	})
}

func TestJobsRepository_Get(t *testing.T) {
	t.Run("Create 3 job and Get the second one", func(t *testing.T) {
		prg := getTestProgram(t)
		testedRepo := NewJobsRepository()
		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"

		_, err := testedRepo.CreateWithID(prg, givenID1)
		assert.NoError(t, err)
		dlg, err := testedRepo.CreateWithID(prg, givenID2)
		assert.NoError(t, err)
		_, err = testedRepo.CreateWithID(prg, givenID3)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.jobs, 3)

		retrievedDlg, ok := testedRepo.Get(givenID2)
		assert.True(t, ok)
		assert.Equal(t, dlg, retrievedDlg)
	})

	t.Run("Get a job that does not exist", func(t *testing.T) {
		testedRepo := NewJobsRepository()

		_, ok := testedRepo.Get("doesNotExist")
		assert.False(t, ok)
	})
}

func TestJobsRepository_GetAll(t *testing.T) {
	t.Run("Create 3 job and Get all of them", func(t *testing.T) {
		prg := getTestProgram(t)
		testedRepo := NewJobsRepository()
		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"

		dlg1, err := testedRepo.CreateWithID(prg, givenID1)
		assert.NoError(t, err)
		dlg2, err := testedRepo.CreateWithID(prg, givenID2)
		assert.NoError(t, err)
		dlg3, err := testedRepo.CreateWithID(prg, givenID3)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.jobs, 3)

		retrievedDlgs := testedRepo.GetAll()
		assert.NotNil(t, retrievedDlgs)
		assert.Len(t, retrievedDlgs, 3)
		assert.Contains(t, retrievedDlgs, dlg1)
		assert.Contains(t, retrievedDlgs, dlg2)
		assert.Contains(t, retrievedDlgs, dlg3)
	})

	t.Run("Get all jobs of an empty repo", func(t *testing.T) {
		testedRepo := NewJobsRepository()

		retrievedDlgs := testedRepo.GetAll()
		assert.NotNil(t, retrievedDlgs)
		assert.Len(t, retrievedDlgs, 0)
	})
}

func TestJobsRepository_Delete(t *testing.T) {
	t.Run("Create 3 job and delete the second one", func(t *testing.T) {
		prg := getTestProgram(t)
		testedRepo := NewJobsRepository()
		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"

		_, err := testedRepo.CreateWithID(prg, givenID1)
		assert.NoError(t, err)
		_, err = testedRepo.CreateWithID(prg, givenID2)
		assert.NoError(t, err)
		_, err = testedRepo.CreateWithID(prg, givenID3)
		assert.NoError(t, err)
		assert.Len(t, testedRepo.jobs, 3)

		ok := testedRepo.Delete(givenID2)
		assert.True(t, ok)
		assert.Len(t, testedRepo.jobs, 2)
		assert.Contains(t, testedRepo.jobs, givenID1)
		assert.NotContains(t, testedRepo.jobs, givenID2)
		assert.Contains(t, testedRepo.jobs, givenID3)
	})

	t.Run("Delete a finished job (no interruption)", func(t *testing.T) {
		prg := getTestProgram(t)
		testedRepo := NewJobsRepository()
		const givenID = "givenID"

		dlg, err := testedRepo.CreateWithID(prg, givenID)
		dlg.Status = core.DelugeDoneError
		assert.NoError(t, err)
		assert.Len(t, testedRepo.jobs, 1)

		ok := testedRepo.Delete(givenID)
		assert.True(t, ok)
		assert.Len(t, testedRepo.jobs, 0)
		assert.Equal(t, core.DelugeDoneError, dlg.Status)
	})

	t.Run("Delete an unfinished job (with interruption)", func(t *testing.T) {
		prg := getTestProgram(t)
		testedRepo := NewJobsRepository()
		const givenID = "givenID"

		dlg, err := testedRepo.CreateWithID(prg, givenID)
		dlg.Status = core.DelugeVirgin
		assert.NoError(t, err)
		assert.Len(t, testedRepo.jobs, 1)

		ok := testedRepo.Delete(givenID)
		assert.True(t, ok)
		assert.Len(t, testedRepo.jobs, 0)
		assert.Equal(t, core.DelugeInterrupted, dlg.Status)
	})

	t.Run("Delete a job that does not exist", func(t *testing.T) {
		testedRepo := NewJobsRepository()

		ok := testedRepo.Delete("doesNotExist")
		assert.False(t, ok)
	})
}

// TestJobsRepository_Race is only meant to be run with the go race detector tool
func TestJobsRepository_Race(t *testing.T) {
	testedRepo := NewJobsRepository()
	const givenID1 = "givenID1"
	const givenID2 = "givenID2"

	go func() {
		prg := getTestProgram(t)

		dlg, err := testedRepo.CreateWithID(prg, givenID1)
		assert.NoError(t, err)
		assert.Equal(t, dlg.ID, givenID1)
	}()

	go func() {
		prg := getTestProgram(t)

		dlg, err := testedRepo.CreateWithID(prg, givenID2)
		assert.NoError(t, err)
		assert.Equal(t, dlg.ID, givenID2)
	}()

	go func() {
		testedRepo.Delete(givenID1)
	}()

	go func() {
		testedRepo.GetAll()
	}()

	go func() {
		testedRepo.Get(givenID1)
	}()
}

func getTestProgram(t testing.TB) *ast.Program {
	l := lexer.New(`
	deluge("Some name", "200ms", {
		"myScenario": {
			"concurrent": 100,
			"delay": "100ms"
		}
	});

	scenario("myScenario", "My scenario", function () {
	});`)
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		core.PrintParserErrors(p.Errors())
		t.Fatal("Parsing error(s)")
	}

	return program
}

func getTestProgramWithError(t testing.TB) *ast.Program {
	l := lexer.New(`
	deluge();`)
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		core.PrintParserErrors(p.Errors())
		t.Fatal("Parsing error(s)")
	}

	return program
}
