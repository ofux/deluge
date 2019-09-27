package api

import (
	"bytes"
	"encoding/json"
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/parser"
	"github.com/ofux/deluge/repo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestJobsWorkerHandler_CreateJob(t *testing.T) {
	const dlgName = "Deluge name"
	const scenarioKey = "myScenario"
	const scenarioName = "My scenario"

	var router = NewRouter(NewJobsWorkerHandler())

	t.Run("Create a job", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()
		w := httptest.NewRecorder()

		r := httptest.NewRequest("POST", "http://example.com/v1/jobs", strings.NewReader(`
			deluge("`+dlgName+`", "200ms", {
				"myScenario": {
					"concurrent": 10,
					"delay": "100ms"
				}
			});

			scenario("`+scenarioKey+`", "`+scenarioName+`", function () {

			});
		`))
		router.ServeHTTP(w, r)

		assert.Equal(t, w.Code, http.StatusAccepted)
		dlg := deserializeDeluge(t, w.Body)
		assert.Equal(t, dlg.DelugeName, dlgName)
		assert.True(t, dlg.Status == JobVirgin || dlg.Status == JobInProgress)
		assert.Len(t, dlg.Scenarios, 1)
		assert.Contains(t, dlg.Scenarios, scenarioKey)
		assert.Equal(t, dlg.Scenarios[scenarioKey].Name, scenarioName)
		assert.Len(t, dlg.Scenarios[scenarioKey].Errors, 0)
		assert.True(t, dlg.Scenarios[scenarioKey].Status == JobScenarioVirgin || dlg.Scenarios[scenarioKey].Status == JobScenarioInProgress)
	})

	t.Run("Create a job with syntax error in script", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()
		w := httptest.NewRecorder()

		r := httptest.NewRequest("POST", "http://example.com/v1/jobs", strings.NewReader(`
			deluge("`+dlgName+`", "200ms", {
			};
		`))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		errDto := deserializeError(t, w.Body)
		assert.Equal(t, "Syntax error:\n\texpected next token to be ), got ; instead (line 3, col 5)\n", errDto.Error)
	})

	t.Run("Create a job without body", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()
		w := httptest.NewRecorder()

		r := httptest.NewRequest("POST", "http://example.com/v1/jobs", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		errDto := deserializeError(t, w.Body)
		assert.Equal(t, "Missing body", errDto.Error)
	})

	t.Run("Create a job with a specific ID", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()
		w := httptest.NewRecorder()

		const dlgID = "foo"
		r := httptest.NewRequest("POST", "http://example.com/v1/jobs?id="+dlgID, strings.NewReader(`
			deluge("`+dlgName+`", "200ms", {
				"myScenario": {
					"concurrent": 10,
					"delay": "100ms"
				}
			});

			scenario("`+scenarioKey+`", "`+scenarioName+`", function () {

			});
		`))
		router.ServeHTTP(w, r)

		assert.Equal(t, w.Code, http.StatusAccepted)
		dlg := deserializeDeluge(t, w.Body)
		assert.Equal(t, dlg.DelugeID, dlgID)
		assert.Equal(t, dlg.DelugeName, dlgName)
		assert.True(t, dlg.Status == JobVirgin || dlg.Status == JobInProgress)
		assert.Len(t, dlg.Scenarios, 1)
		assert.Contains(t, dlg.Scenarios, scenarioKey)
		assert.Equal(t, dlg.Scenarios[scenarioKey].Name, scenarioName)
		assert.Len(t, dlg.Scenarios[scenarioKey].Errors, 0)
		assert.True(t, dlg.Scenarios[scenarioKey].Status == JobScenarioVirgin || dlg.Scenarios[scenarioKey].Status == JobScenarioInProgress)
	})

	t.Run("Create a job with an existing ID", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()
		repo.Jobs = repo.NewJobsRepository()

		const dlgID = "bar"
		const body = `
			deluge("` + dlgName + `", "200ms", {
				"myScenario": {
					"concurrent": 10,
					"delay": "100ms"
				}
			});

			scenario("` + scenarioKey + `", "` + scenarioName + `", function () {

			});
		`
		r := httptest.NewRequest("POST", "http://example.com/v1/jobs?id="+dlgID, strings.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		assert.Equal(t, w.Code, http.StatusAccepted)
		dlg := deserializeDeluge(t, w.Body)
		assert.Equal(t, dlg.DelugeID, dlgID)

		r = httptest.NewRequest("POST", "http://example.com/v1/jobs?id="+dlgID, strings.NewReader(body))
		w = httptest.NewRecorder()
		router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		errDto := deserializeError(t, w.Body)
		assert.Equal(t, "Cannot create job with id 'bar'. A job with this id already exists.", errDto.Error)
	})

	t.Run("Create a job with a webhook", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()
		w := httptest.NewRecorder()

		webhook := make(chan struct{})
		webhookSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.FormValue("job_id")
			assert.Len(t, id, 36)
			close(webhook)
		}))

		r := httptest.NewRequest("POST", "http://example.com/v1/jobs?webhook="+webhookSrv.URL, strings.NewReader(`
			deluge("`+dlgName+`", "200ms", {
				"myScenario": {
					"concurrent": 10,
					"delay": "100ms"
				}
			});

			scenario("`+scenarioKey+`", "`+scenarioName+`", function () {

			});
		`))
		router.ServeHTTP(w, r)

		assert.Equal(t, w.Code, http.StatusAccepted)
		dlg := deserializeDeluge(t, w.Body)
		assert.Equal(t, dlg.DelugeName, dlgName)
		assert.True(t, dlg.Status == JobVirgin || dlg.Status == JobInProgress)
		assert.Len(t, dlg.Scenarios, 1)

		assert.True(t, isChanClosed(webhook, 50, 100*time.Millisecond))
	})

	t.Run("Create a job with a bad webhook URL", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()
		w := httptest.NewRecorder()

		r := httptest.NewRequest("POST", "http://example.com/v1/jobs?webhook=not%5Ea%7Cvalid%3Eurl", strings.NewReader(`
			deluge("`+dlgName+`", "200ms", {
				"myScenario": {
					"concurrent": 10,
					"delay": "100ms"
				}
			});

			scenario("`+scenarioKey+`", "`+scenarioName+`", function () {

			});
		`))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		errDto := deserializeError(t, w.Body)
		assert.Equal(t, "parse not^a|valid>url: invalid URI for request", errDto.Error)
	})
}

func TestJobsWorkerHandler_GetJob(t *testing.T) {

	var router = NewRouter(NewJobsWorkerHandler())

	t.Run("Get a job", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()

		const givenID = "givenID"
		repo.Jobs.CreateWithID(getTestProgram(t), givenID)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://example.com/v1/jobs/"+givenID, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		dlg := deserializeDeluge(t, w.Body)
		assert.Equal(t, dlg.DelugeID, givenID)
		assert.Equal(t, dlg.DelugeName, "Some name")
	})

	t.Run("Get a job that does not exist", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()

		repo.Jobs.CreateWithID(getTestProgram(t), "givenID")

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://example.com/v1/jobs/badID", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestJobsWorkerHandler_GetAllJobs(t *testing.T) {

	var router = NewRouter(NewJobsWorkerHandler())

	t.Run("Get all jobs", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()

		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"
		repo.Jobs.CreateWithID(getTestProgram(t), givenID1)
		repo.Jobs.CreateWithID(getTestProgram(t), givenID2)
		repo.Jobs.CreateWithID(getTestProgram(t), givenID3)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://example.com/v1/jobs", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		dlgs := deserializeArrayOfDeluges(t, w.Body)
		assert.Len(t, dlgs, 3)
	})

	t.Run("Get all jobs of an empty repo", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://example.com/v1/jobs", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		dlgs := deserializeArrayOfDeluges(t, w.Body)
		assert.Len(t, dlgs, 0)
	})
}

func TestJobsOrchestratorHandler_DeleteJob(t *testing.T) {
	var router = NewRouter(NewJobsWorkerHandler())

	t.Run("Delete a finished job", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()

		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"
		repo.Jobs.CreateWithID(getTestProgram(t), givenID1)
		repo.Jobs.CreateWithID(getTestProgram(t), givenID2)
		repo.Jobs.CreateWithID(getTestProgram(t), givenID3)
		dlg, _ := repo.Jobs.Get(givenID2)
		dlg.Status = core.DelugeDoneSuccess

		w := httptest.NewRecorder()
		r := httptest.NewRequest("DELETE", "http://example.com/v1/jobs/"+givenID2, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, core.DelugeDoneSuccess, dlg.Status)

		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "http://example.com/v1/jobs", nil)
		router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		dlgs := deserializeArrayOfDeluges(t, w.Body)
		assert.Len(t, dlgs, 2)
	})

	t.Run("Delete a non-existing job", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()

		w := httptest.NewRecorder()
		r := httptest.NewRequest("DELETE", "http://example.com/v1/jobs/badID", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Delete an unfinished job", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()

		const givenID1 = "givenID1"
		const givenID2 = "givenID2"
		const givenID3 = "givenID3"
		repo.Jobs.CreateWithID(getTestProgram(t), givenID1)
		interruptedDlg, _ := repo.Jobs.CreateWithID(getTestProgram(t), givenID2)
		repo.Jobs.CreateWithID(getTestProgram(t), givenID3)
		dlg, _ := repo.Jobs.Get(givenID2)
		dlg.Status = core.DelugeInProgress

		w := httptest.NewRecorder()
		r := httptest.NewRequest("DELETE", "http://example.com/v1/jobs/"+givenID2, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, core.DelugeInterrupted, interruptedDlg.Status)

		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "http://example.com/v1/jobs", nil)
		router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		dlgs := deserializeArrayOfDeluges(t, w.Body)
		assert.Len(t, dlgs, 2)
	})
}

func TestJobsOrchestratorHandler_InterruptJob(t *testing.T) {
	var router = NewRouter(NewJobsWorkerHandler())

	t.Run("Interrupt a successful job", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()

		const givenID = "givenID"
		repo.Jobs.CreateWithID(getTestProgram(t), givenID)
		dlg, _ := repo.Jobs.Get(givenID)
		dlg.Status = core.DelugeDoneSuccess

		w := httptest.NewRecorder()
		r := httptest.NewRequest("PUT", "http://example.com/v1/jobs/interrupt/"+givenID, nil)
		router.ServeHTTP(w, r)

		// Check status hasn't changed
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, core.DelugeDoneSuccess, dlg.Status)

		// Check deluge is still there with the same status
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "http://example.com/v1/jobs/"+givenID, nil)
		router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		dlgGet := deserializeDeluge(t, w.Body)
		assert.Equal(t, givenID, dlgGet.DelugeID)
		assert.Equal(t, JobDoneSuccess, dlgGet.Status)
	})

	t.Run("Interrupt a job done with errors", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()

		const givenID = "givenID"
		repo.Jobs.CreateWithID(getTestProgram(t), givenID)
		dlg, _ := repo.Jobs.Get(givenID)
		dlg.Status = core.DelugeDoneError

		w := httptest.NewRecorder()
		r := httptest.NewRequest("PUT", "http://example.com/v1/jobs/interrupt/"+givenID, nil)
		router.ServeHTTP(w, r)

		// Check status hasn't changed
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, core.DelugeDoneError, dlg.Status)

		// Check deluge is still there with the same status
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "http://example.com/v1/jobs/"+givenID, nil)
		router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		dlgGet := deserializeDeluge(t, w.Body)
		assert.Equal(t, givenID, dlgGet.DelugeID)
		assert.Equal(t, JobDoneError, dlgGet.Status)
	})

	t.Run("Interrupt a job in progress", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()

		const givenID = "givenID"
		repo.Jobs.CreateWithID(getTestProgram(t), givenID)
		dlg, _ := repo.Jobs.Get(givenID)
		dlg.Status = core.DelugeInProgress

		w := httptest.NewRecorder()
		r := httptest.NewRequest("PUT", "http://example.com/v1/jobs/interrupt/"+givenID, nil)
		router.ServeHTTP(w, r)

		// Check deluge status has been changed
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, core.DelugeInterrupted, dlg.Status)

		// Check deluge is still there with status interrupted
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "http://example.com/v1/jobs/"+givenID, nil)
		router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		dlgGet := deserializeDeluge(t, w.Body)
		assert.Equal(t, givenID, dlgGet.DelugeID)
		assert.Equal(t, JobInterrupted, dlgGet.Status)
	})

	t.Run("Interrupt a non-existing job", func(t *testing.T) {
		repo.Jobs = repo.NewJobsRepository()

		w := httptest.NewRecorder()
		r := httptest.NewRequest("PUT", "http://example.com/v1/jobs/interrupt/badID", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func isChanClosed(ch chan struct{}, maxIterations int, iterationTime time.Duration) bool {
	for i := 0; i < maxIterations; i++ {
		select {
		case _, ok := <-ch:
			if !ok {
				return true
			}
		default:
			time.Sleep(iterationTime)
		}
	}
	return false
}

func deserializeArrayOfDeluges(t *testing.T, body *bytes.Buffer) []*Job {
	p := make([]byte, body.Len())
	if _, err := body.Read(p); err != nil {
		t.Fatalf("Could not read body")
		return nil
	}
	dlgs := make([]*Job, 0)
	if err := json.Unmarshal(p, &dlgs); err != nil {
		t.Fatalf("Could not deserialize array of Deluges out of %s", string(p))
		return nil
	}
	return dlgs
}

func deserializeDeluge(t *testing.T, body *bytes.Buffer) *Job {
	p := make([]byte, body.Len())
	if _, err := body.Read(p); err != nil {
		t.Fatalf("Could not read body")
		return nil
	}
	dlg := &Job{}
	if err := json.Unmarshal(p, dlg); err != nil {
		t.Fatalf("Could not deserialize Deluge out of %s", string(p))
		return nil
	}
	return dlg
}

func deserializeError(t *testing.T, body *bytes.Buffer) *Error {
	p := make([]byte, body.Len())
	if _, err := body.Read(p); err != nil {
		t.Fatalf("Could not read body")
		return nil
	}
	errDto := &Error{}
	if err := json.Unmarshal(p, errDto); err != nil {
		t.Fatalf("Could not deserialize Error out of %s", string(p))
		return nil
	}
	return errDto
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
