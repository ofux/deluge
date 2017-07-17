package api

import (
	"github.com/gorilla/mux"
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/parser"
	"github.com/ofux/deluge/repo"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
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
	// Delete one Job
	routes = append(routes, Route{
		Name:        "Delete a job",
		Method:      http.MethodDelete,
		Pattern:     "/{id}",
		HandlerFunc: jobsHandler.DeleteJob,
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
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		SendJSONError(w, "Missing body", http.StatusBadRequest)
		return
	}

	l := lexer.New(string(body))
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		SendJSONError(w, core.SPrintParserErrors(p.Errors()), http.StatusBadRequest)
		return
	}

	jobID := r.FormValue("id")
	var dlg *core.Deluge
	if jobID == "" {
		dlg, err = repo.Jobs.Create(program)
	} else {
		dlg, err = repo.Jobs.CreateWithID(program, jobID)
	}
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	wURLParam := r.FormValue("webhook")
	if wURLParam == "" {
		dlg.Run()
	} else {
		wURL, err := url.ParseRequestURI(wURLParam)
		if err != nil {
			SendJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		q := wURL.Query()
		q.Set("job_id", dlg.ID)
		wURL.RawQuery = q.Encode()
		go func() {
			<-dlg.Run()
			resp, err := http.Get(wURL.String())
			if err != nil {
				log.Warnf("Error calling webhook (%s): %v", wURL.String(), err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				log.Warnf("The webhook (%s) responded with status: %d (%s)", wURL.String(), resp.StatusCode, resp.Status)
			}
		}()
	}

	SendJSONWithHTTPCode(w, mapDeluge(dlg), http.StatusAccepted)
}

func (d *JobsWorkerHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	dlg, ok := repo.Jobs.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	SendJSONWithHTTPCode(w, mapDeluge(dlg), http.StatusOK)
}

func (d *JobsWorkerHandler) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	dlgs := repo.Jobs.GetAll()
	dlgsDTO := make([]*DelugeLite, 0, len(dlgs))
	for _, dlg := range dlgs {
		dlgsDTO = append(dlgsDTO, mapDelugeLite(dlg))
	}

	SendJSONWithHTTPCode(w, dlgsDTO, http.StatusOK)
}

func (d *JobsWorkerHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	ok := repo.Jobs.Delete(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (d *JobsWorkerHandler) InterruptJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	dlg, ok := repo.Jobs.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dlg.Interrupt()
	w.WriteHeader(http.StatusOK)
}
