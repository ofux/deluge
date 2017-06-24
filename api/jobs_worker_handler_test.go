package api

import (
	"bytes"
	"encoding/json"
	"github.com/ofux/deluge/api/dto"
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

	t.Run("Create a job", func(t *testing.T) {
		w := httptest.NewRecorder()
		workerHandler := &JobsWorkerHandler{}

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
		workerHandler.CreateJob(w, r)

		assert.Equal(t, w.Code, http.StatusAccepted)
		dlg := deserializeDeluge(t, w.Body)
		assert.Equal(t, dlg.Name, dlgName)
		assert.True(t, dlg.Status == dto.DelugeVirgin || dlg.Status == dto.DelugeInProgress)
		assert.Len(t, dlg.Scenarios, 1)
		assert.Contains(t, dlg.Scenarios, scenarioKey)
		assert.Equal(t, dlg.Scenarios[scenarioKey].Name, scenarioName)
		assert.Len(t, dlg.Scenarios[scenarioKey].Errors, 0)
		assert.True(t, dlg.Scenarios[scenarioKey].Status == dto.ScenarioVirgin || dlg.Scenarios[scenarioKey].Status == dto.ScenarioInProgress)
	})

	t.Run("Create a job with syntax error in script", func(t *testing.T) {

		w := httptest.NewRecorder()
		workerHandler := &JobsWorkerHandler{}

		r := httptest.NewRequest("POST", "http://example.com/v1/jobs", strings.NewReader(`
			deluge("`+dlgName+`", "200ms", {
			};
		`))
		workerHandler.CreateJob(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		errDto := deserializeError(t, w.Body)
		assert.Equal(t, "Syntax error:\n\texpected next token to be ), got ; instead (line 3, col 5)\n", errDto.Error)
	})

	t.Run("Create a job without body", func(t *testing.T) {

		w := httptest.NewRecorder()
		workerHandler := &JobsWorkerHandler{}

		r := httptest.NewRequest("POST", "http://example.com/v1/jobs", nil)
		workerHandler.CreateJob(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		errDto := deserializeError(t, w.Body)
		assert.Equal(t, "Missing body", errDto.Error)
	})

	t.Run("Create a job with a specific ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		workerHandler := &JobsWorkerHandler{}

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
		workerHandler.CreateJob(w, r)

		assert.Equal(t, w.Code, http.StatusAccepted)
		dlg := deserializeDeluge(t, w.Body)
		assert.Equal(t, dlg.ID, dlgID)
		assert.Equal(t, dlg.Name, dlgName)
		assert.True(t, dlg.Status == dto.DelugeVirgin || dlg.Status == dto.DelugeInProgress)
		assert.Len(t, dlg.Scenarios, 1)
		assert.Contains(t, dlg.Scenarios, scenarioKey)
		assert.Equal(t, dlg.Scenarios[scenarioKey].Name, scenarioName)
		assert.Len(t, dlg.Scenarios[scenarioKey].Errors, 0)
		assert.True(t, dlg.Scenarios[scenarioKey].Status == dto.ScenarioVirgin || dlg.Scenarios[scenarioKey].Status == dto.ScenarioInProgress)
	})

	t.Run("Create a job with an existing ID", func(t *testing.T) {
		workerHandler := &JobsWorkerHandler{}

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
		workerHandler.CreateJob(w, r)
		assert.Equal(t, w.Code, http.StatusAccepted)
		dlg := deserializeDeluge(t, w.Body)
		assert.Equal(t, dlg.ID, dlgID)

		r = httptest.NewRequest("POST", "http://example.com/v1/jobs?id="+dlgID, strings.NewReader(body))
		w = httptest.NewRecorder()
		workerHandler.CreateJob(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		errDto := deserializeError(t, w.Body)
		assert.Equal(t, "Cannot create job with id 'bar'. A job with this id already exists.", errDto.Error)
	})

	t.Run("Create a job with a webhook", func(t *testing.T) {
		w := httptest.NewRecorder()
		workerHandler := &JobsWorkerHandler{}

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
		workerHandler.CreateJob(w, r)

		assert.Equal(t, w.Code, http.StatusAccepted)
		dlg := deserializeDeluge(t, w.Body)
		assert.Equal(t, dlg.Name, dlgName)
		assert.True(t, dlg.Status == dto.DelugeVirgin || dlg.Status == dto.DelugeInProgress)
		assert.Len(t, dlg.Scenarios, 1)

		assert.True(t, isChanClosed(webhook, 50, 100*time.Millisecond))
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

func deserializeDeluge(t *testing.T, body *bytes.Buffer) *dto.Deluge {
	p := make([]byte, body.Len())
	if _, err := body.Read(p); err != nil {
		t.Fatalf("Could not read body")
		return nil
	}
	dlg := &dto.Deluge{}
	if err := json.Unmarshal(p, dlg); err != nil {
		t.Fatalf("Could not deserialize Deluge out of %s", string(p))
		return nil
	}
	return dlg
}

func deserializeError(t *testing.T, body *bytes.Buffer) *dto.Error {
	p := make([]byte, body.Len())
	if _, err := body.Read(p); err != nil {
		t.Fatalf("Could not read body")
		return nil
	}
	errDto := &dto.Error{}
	if err := json.Unmarshal(p, errDto); err != nil {
		t.Fatalf("Could not deserialize Error out of %s", string(p))
		return nil
	}
	return errDto
}
