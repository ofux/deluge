package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ofux/deluge/repov2"
	"github.com/ofux/deluge/worker"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"net/url"
)

// JobsHandler handles requests for 'jobs' resource
type JobsHandler struct {
	routes []Route
}

func (d *JobsHandler) GetBasePath() string {
	return "/v1/jobs"
}

func (d *JobsHandler) GetRoutes() []Route {
	return d.routes
}

// NewJobsWorkerHandler adds handlers to handle jobs
func NewJobsWorkerHandler() *JobsHandler {
	jobsHandler := &JobsHandler{}

	// build routes
	routes := []Route{}
	// Create a Job
	routes = append(routes, Route{
		Name:        "Create and run a job",
		Method:      http.MethodPost,
		Pattern:     "",
		HandlerFunc: jobsHandler.CreateJob,
	})
	// Get one Job
	routes = append(routes, Route{
		Name:        "Get a job",
		Method:      http.MethodGet,
		Pattern:     "/{id}",
		HandlerFunc: jobsHandler.GetJob,
	})
	// Get all Jobs
	routes = append(routes, Route{
		Name:        "Get all jobs",
		Method:      http.MethodGet,
		Pattern:     "",
		HandlerFunc: jobsHandler.GetAllJobs,
	})
	// Interrupt a Job
	routes = append(routes, Route{
		Name:        "Interrupt a job",
		Method:      http.MethodPut,
		Pattern:     "/interrupt/{id}",
		HandlerFunc: jobsHandler.InterruptJob,
	})

	jobsHandler.routes = routes

	return jobsHandler
}

func (d *JobsHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var job JobCreation
	if ok := GetJSONBody(w, r, &job); !ok {
		return
	}

	_, exists := repov2.Instance.GetDeluge(job.DelugeID)
	if !exists {
		SendJSONError(w, fmt.Sprintf("Deluge with ID '%s' does not exist.", job.DelugeID), http.StatusNotFound)
		return
	}

	jobID := uuid.NewV4().String()

	webhook := ""
	if job.Webhook != "" {
		wURL, err := url.ParseRequestURI(job.Webhook)
		if err != nil {
			SendJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		q := wURL.Query()
		q.Set("job_id", jobID)
		wURL.RawQuery = q.Encode()
		webhook = wURL.String()
	}

	jobShell := &repov2.PersistedJobShell{
		ID:       jobID,
		DelugeID: job.DelugeID,
		Webhook:  webhook,
	}
	err := repov2.Instance.SaveJobShell(jobShell)
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = startJob(jobShell)
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respDTO := &JobMetadata{
		ID:       jobShell.ID,
		DelugeID: jobShell.DelugeID,
		Webhook:  jobShell.Webhook,
	}

	SendJSONWithHTTPCode(w, respDTO, http.StatusAccepted)
}

func startJob(jobShell *repov2.PersistedJobShell) error {
	err := worker.GetManager().CreateAll(&worker.JobShell{
		ID:       jobShell.ID,
		DelugeID: jobShell.DelugeID,
	})
	if err != nil {
		return err
	}

	return worker.GetManager().StartAll(&worker.JobShell{
		ID:       jobShell.ID,
		DelugeID: jobShell.DelugeID,
	})
}

func (d *JobsHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	job, ok := repov2.Instance.GetJobShell(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	partialContent := false
	deluge, ok := repov2.Instance.GetDeluge(job.DelugeID)
	if !ok {
		partialContent = true
	}

	scenarios := repov2.Instance.GetDelugeScenarios(deluge.ScenarioIDs)
	if len(scenarios) == 0 {
		partialContent = true
	}

	reports := repov2.Instance.GetJobWorkerReports(id)
	if len(reports) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jobReport, err := mapDeluge(job, deluge, scenarios, reports)
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !jobReport.Status.IsEnd() {
		partialContent = true
	}

	if partialContent {
		SendJSONWithHTTPCode(w, jobReport, http.StatusPartialContent)
	} else {
		SendJSONWithHTTPCode(w, jobReport, http.StatusOK)
	}
}

func (d *JobsHandler) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	jobs := repov2.Instance.GetAllJobShell()
	dlgsDTO := make([]JobMetadata, 0, len(jobs))
	for _, job := range jobs {
		dlgsDTO = append(dlgsDTO, JobMetadata{
			ID:       job.ID,
			DelugeID: job.DelugeID,
			Webhook:  job.Webhook,
		})
	}

	SendJSONWithHTTPCode(w, dlgsDTO, http.StatusOK)
}

func (d *JobsHandler) InterruptJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	err := worker.GetManager().InterruptAll(id)
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
