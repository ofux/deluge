package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/repo"
	"github.com/ofux/deluge/repov2"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

// JobsWorkerHandler handles requests for 'jobs' resource as a worker
type JobsWorkerHandler struct {
	routes []Route
}

func (d *JobsWorkerHandler) GetBasePath() string {
	return "/v1/jobs"
}

func (d *JobsWorkerHandler) GetRoutes() []Route {
	return d.routes
}

// NewJobsWorkerHandler adds handlers to handle jobs as a worker
func NewJobsWorkerHandler() *JobsWorkerHandler {
	jobsHandler := &JobsWorkerHandler{}

	// build routes
	routes := []Route{}
	// Create a Job
	routes = append(routes, Route{
		Name:        "Creates and runs a job",
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
	// Start a Job
	routes = append(routes, Route{
		Name:        "Interrupt a job",
		Method:      http.MethodPut,
		Pattern:     "/start/{id}",
		HandlerFunc: jobsHandler.StartJob,
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

func (d *JobsWorkerHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var job JobCreation
	if ok := GetJSONBody(w, r, &job); !ok {
		return
	}

	_, exists := repov2.Instance.GetDeluge(job.DelugeID)
	if !exists {
		SendJSONError(w, fmt.Sprintf("Deluge with ID '%s' does not exist.", job.DelugeID), http.StatusBadRequest)
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

	err := repov2.Instance.SaveJobShell(&repov2.PersistedJobShell{
		ID:       jobID,
		DelugeID: job.DelugeID,
		Webhook:  webhook,
	})
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respDTO := &JobLite{
		ID:       jobID,
		DelugeID: job.DelugeID,
		Status:   JobVirgin,
	}

	SendJSONWithHTTPCode(w, respDTO, http.StatusAccepted)
}

func (d *JobsWorkerHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	job, ok := repo.Jobs.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	SendJSONWithHTTPCode(w, mapDeluge(job.RunnableDeluge), http.StatusOK)
}

func (d *JobsWorkerHandler) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	jobs := repo.Jobs.GetAll()
	dlgsDTO := make([]*JobLite, 0, len(jobs))
	for _, job := range jobs {
		dlgsDTO = append(dlgsDTO, mapDelugeLite(job.RunnableDeluge))
	}

	SendJSONWithHTTPCode(w, dlgsDTO, http.StatusOK)
}

func (d *JobsWorkerHandler) StartJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	jobShell, ok := repov2.Instance.GetJobShell(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dlg, err := core.NewRunnableDeluge(jobShell.DelugeID)
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repo.Jobs.Store(&repo.RunningJob{
		ID:             jobShell.ID,
		RunnableDeluge: dlg,
	})

	if jobShell.Webhook == "" {
		dlg.Run()
	} else {
		go func() {
			<-dlg.Run()
			resp, err := http.Get(jobShell.Webhook)
			if err != nil {
				log.Warnf("Error calling webhook (%s): %v", jobShell.Webhook, err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				log.Warnf("The webhook (%s) responded with status: %d (%s)", jobShell.Webhook, resp.StatusCode, resp.Status)
			}
		}()
	}

	w.WriteHeader(http.StatusAccepted)
}

func (d *JobsWorkerHandler) InterruptJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	job, ok := repo.Jobs.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	job.RunnableDeluge.Interrupt()
	w.WriteHeader(http.StatusOK)
}
