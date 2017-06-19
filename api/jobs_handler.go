package api

import (
	"github.com/gorilla/mux"
	"github.com/ofux/deluge-dsl/lexer"
	"github.com/ofux/deluge-dsl/parser"
	"github.com/ofux/deluge/api/dto"
	"github.com/ofux/deluge/api/repo"
	"github.com/ofux/deluge/core"
	"io/ioutil"
	"net/http"
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

// NewTaskController creates a new task controller to manage tasks
func NewDelugeHandler() *JobsHandler {
	jobsHandler := &JobsHandler{}

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

	jobsHandler.routes = routes

	return jobsHandler
}

func (d *JobsHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		SendJSONError(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	l := lexer.New(string(body))
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		SendJSONError(w, core.SPrintParserErrors(p.Errors()), http.StatusBadRequest)
		return
	}

	dlg := repo.Jobs.Create(program)
	go dlg.Run()

	SendJSONWithHTTPCode(w, dto.MapDeluge(dlg), http.StatusAccepted)
}

func (d *JobsHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	dlg, ok := repo.Jobs.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	SendJSONWithHTTPCode(w, dto.MapDeluge(dlg), http.StatusOK)
}

func (d *JobsHandler) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	dlgs := repo.Jobs.GetAll()
	dlgsDTO := make([]*dto.DelugeLite, 0, len(dlgs))
	for _, dlg := range dlgs {
		dlgsDTO = append(dlgsDTO, dto.MapDelugeLite(dlg))
	}

	SendJSONWithHTTPCode(w, dlgsDTO, http.StatusOK)
}

func (d *JobsHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	ok := repo.Jobs.Delete(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
