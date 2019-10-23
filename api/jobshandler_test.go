package api

import (
	"encoding/json"
	"errors"
	"github.com/ofux/deluge/core/status"
	"github.com/ofux/deluge/repov2"
	"github.com/ofux/deluge/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJobsHandler_Create(t *testing.T) {
	const scenarioKey = "myScenario"
	const delugeKey = "myDeluge"

	var router = NewRouter(NewJobHandler())

	t.Run("Create a valid job", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		createScenario(t, scenarioKey, "My scenario")
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		var body = `{
			"delugeId": "` + delugeKey + `",
			"webhook": "http://someurl.com"
		}`

		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/jobs", strings.NewReader(body))
		router.ServeHTTP(w, r)

		require.Equal(t, http.StatusAccepted, w.Code)

		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		var response JobMetadata
		err = json.Unmarshal(bbody, &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.ID)
		assert.Equal(t, delugeKey, response.DelugeID)
		assert.Equal(t, "http://someurl.com?job_id="+response.ID, response.Webhook)

		job, ok := repov2.Instance.GetJobShell(response.ID)
		require.True(t, ok)
		require.NotNil(t, job)
		assert.Equal(t, job.ID, response.ID)
		assert.Equal(t, delugeKey, job.DelugeID)
		assert.Equal(t, "http://someurl.com?job_id="+response.ID, job.Webhook)
	})

	t.Run("Create a valid job with minimal info", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		createScenario(t, scenarioKey, "My scenario")
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		var body = `{
			"delugeId": "` + delugeKey + `"
		}`

		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/jobs", strings.NewReader(body))
		router.ServeHTTP(w, r)

		assert.Equal(t, w.Code, http.StatusAccepted)

		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		var response JobMetadata
		err = json.Unmarshal(bbody, &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.ID)
		assert.Equal(t, response.DelugeID, delugeKey)
		assert.Equal(t, response.Webhook, "")

		job, ok := repov2.Instance.GetJobShell(response.ID)
		require.True(t, ok)
		require.NotNil(t, job)
		assert.Equal(t, job.ID, response.ID)
		assert.Equal(t, job.DelugeID, delugeKey)
		assert.Equal(t, job.Webhook, "")
	})

	t.Run("Create a job with undefined deluge", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		createScenario(t, scenarioKey, "My scenario")
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		var body = `{
			"delugeId": "IDontExist"
		}`

		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/jobs", strings.NewReader(body))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Create a job and fails to create workers", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = &workerManagerMock{
			CreateAllImpl: func(jobShell *worker.JobShell) error {
				return errors.New("some error")
			},
		}
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		var body = `{
			"delugeId": "` + delugeKey + `"
		}`

		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/jobs", strings.NewReader(body))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Create a job that fails to start", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = &workerManagerMock{
			StartAllImpl: func(jobShell *worker.JobShell) error {
				return errors.New("some error")
			},
		}
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		var body = `{
			"delugeId": "` + delugeKey + `"
		}`

		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/jobs", strings.NewReader(body))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Create a job with invalid webhook", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		createScenario(t, scenarioKey, "My scenario")
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		var body = `{
			"delugeId": "` + delugeKey + `",
			"webhook": "badurl"
		}`

		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/jobs", strings.NewReader(body))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Create malformed job", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		createScenario(t, scenarioKey, "My scenario")
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		var body = `{
			"deluge`
		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/jobs", strings.NewReader(body))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Fails to save job in repository", func(t *testing.T) {
		repov2.Instance = &repoMock{
			InMemoryRepository: *repov2.NewInMemoryRepository(),
			SaveJobShellImpl: func(job *repov2.PersistedJobShell) error {
				return errors.New("some error")
			},
		}
		worker.ManagerInstance = newWorkerManagerMock()
		createScenario(t, scenarioKey, "My scenario")
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		var body = `{
			"delugeId": "` + delugeKey + `"
		}`

		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/jobs", strings.NewReader(body))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Empty body", func(t *testing.T) {
		repov2.Instance = &repoMock{
			InMemoryRepository: *repov2.NewInMemoryRepository(),
		}
		worker.ManagerInstance = newWorkerManagerMock()
		createScenario(t, scenarioKey, "My scenario")
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/jobs", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestJobsHandler_GetByID(t *testing.T) {
	const scenarioKey = "myScenario"
	const delugeKey = "myDeluge"
	const jobKey = "myJob"

	var router = NewRouter(NewJobHandler())

	t.Run("Get an existing job", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		createScenario(t, scenarioKey, "My scenario")
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		createJob(t, jobKey, delugeKey, "")
		createJobReport(t, "workerId", jobKey, status.DelugeDoneSuccess)

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/jobs/"+jobKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		body := string(bbody)
		assert.JSONEq(t, `{"id":"myJob","delugeId":"myDeluge","delugeName":"My deluge","status":"doneSuccess","globalDuration":200000000,"scenarios":{}}`, body)
	})

	t.Run("Get a non-existing job", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		createScenario(t, scenarioKey, "My scenario")
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/jobs/"+jobKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Get an existing job without any report", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		createScenario(t, scenarioKey, "My scenario")
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		createJob(t, jobKey, delugeKey, "")

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/jobs/"+jobKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Get an existing unfinished job", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		createScenario(t, scenarioKey, "My scenario")
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		createJob(t, jobKey, delugeKey, "")
		createJobReport(t, "workerId", jobKey, status.DelugeInProgress)

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/jobs/"+jobKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusPartialContent, w.Code)
		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		body := string(bbody)
		assert.JSONEq(t, `{"id":"myJob","delugeId":"myDeluge","delugeName":"My deluge","status":"inProgress","globalDuration":200000000,"scenarios":{}}`, body)
	})

	t.Run("Get an existing job without scenario definition", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		createDeluge(t, delugeKey, "My deluge", scenarioKey)
		w := httptest.NewRecorder()

		createJob(t, jobKey, delugeKey, "")
		createJobReport(t, "workerId", jobKey, status.DelugeDoneSuccess)

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/jobs/"+jobKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusPartialContent, w.Code)
		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		body := string(bbody)
		assert.JSONEq(t, `{"id":"myJob","delugeId":"myDeluge","delugeName":"My deluge","status":"doneSuccess","globalDuration":200000000,"scenarios":{}}`, body)
	})

	t.Run("Get an existing job without deluge definition", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		createScenario(t, scenarioKey, "My scenario")
		w := httptest.NewRecorder()

		createJob(t, jobKey, delugeKey, "")
		createJobReport(t, "workerId", jobKey, status.DelugeDoneSuccess)

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/jobs/"+jobKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusPartialContent, w.Code)
		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		body := string(bbody)
		assert.JSONEq(t, `{"id":"myJob","delugeId":"myDeluge", "delugeName":"not found","status":"doneSuccess","globalDuration":0,"scenarios":{}}`, body)
	})
}

func TestJobsHandler_GetAll(t *testing.T) {

	const jobKey1 = "myJob1"
	const jobKey2 = "myJob2"

	const delugeKey1 = "myDeluge1"
	const delugeKey2 = "myDeluge2"

	var router = NewRouter(NewJobHandler())

	t.Run("Get all jobs", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		w := httptest.NewRecorder()

		createJob(t, jobKey1, delugeKey1, "")
		createJob(t, jobKey2, delugeKey2, "")

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/jobs", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		body := string(bbody)
		assert.JSONEq(t, `{"elements":[
			{
				"id": "`+jobKey1+`",
				"delugeId": "`+delugeKey1+`",
				"webhook": ""
			},{
				"id": "`+jobKey2+`",
				"delugeId": "`+delugeKey2+`",
				"webhook": ""
			}
		]}`, body)
	})

	t.Run("Get all jobs on empty repository", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = newWorkerManagerMock()
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/jobs", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		body := string(bbody)
		assert.JSONEq(t, `{"elements":[]}`, body)
	})
}

func TestJobsHandler_InterruptJob(t *testing.T) {

	const jobKey = "myJob1"

	var router = NewRouter(NewJobHandler())

	t.Run("Interrupt a job", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = &workerManagerMock{
			InterruptAllImpl: func(jobShellID string) error {
				return nil
			},
		}
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodPut, "http://example.com/v1/jobs/interrupt/"+jobKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusAccepted, w.Code)
	})

	t.Run("Interrupt a job failing", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		worker.ManagerInstance = &workerManagerMock{
			InterruptAllImpl: func(jobShellID string) error {
				return errors.New("some error")
			},
		}
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodPut, "http://example.com/v1/jobs/interrupt/"+jobKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func createJob(t *testing.T, ID, delugeID, webhook string) {
	t.Helper()
	err := repov2.Instance.SaveJobShell(&repov2.PersistedJobShell{
		ID:       ID,
		DelugeID: delugeID,
		Webhook:  webhook,
	})
	require.NoError(t, err)
}

func createJobReport(t *testing.T, workerID, jobID string, status status.DelugeStatus) {
	t.Helper()
	err := repov2.Instance.SaveWorkerReport(&repov2.PersistedWorkerReport{
		WorkerID:  workerID,
		JobID:     jobID,
		Status:    status,
		Scenarios: nil,
	})
	require.NoError(t, err)
}

func (r *repoMock) SaveJobShell(job *repov2.PersistedJobShell) error {
	if r.SaveJobShellImpl == nil {
		return r.InMemoryRepository.SaveJobShell(job)
	}
	return r.SaveJobShellImpl(job)
}

func newWorkerManagerMock() *workerManagerMock {
	return &workerManagerMock{}
}

type workerManagerMock struct {
	CreateAllImpl    func(jobShell *worker.JobShell) error
	StartAllImpl     func(jobShell *worker.JobShell) error
	InterruptAllImpl func(jobShellID string) error
}

func (w *workerManagerMock) CreateAll(jobShell *worker.JobShell) error {
	if w.CreateAllImpl != nil {
		return w.CreateAllImpl(jobShell)
	}
	return nil
}

func (w *workerManagerMock) StartAll(jobShell *worker.JobShell) error {
	if w.StartAllImpl != nil {
		return w.StartAllImpl(jobShell)
	}
	return nil
}

func (w *workerManagerMock) InterruptAll(jobShellID string) error {
	if w.InterruptAllImpl != nil {
		return w.InterruptAllImpl(jobShellID)
	}
	return nil
}
