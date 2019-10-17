package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/repov2"
	"net/http"
)

// ScenarioHandler handles requests for 'scenarios' resource
type ScenarioHandler struct {
	routes []Route
}

func (d *ScenarioHandler) GetBasePath() string {
	return "/v1/scenarios"
}

func (d *ScenarioHandler) GetRoutes() []Route {
	return d.routes
}

// NewScenarioHandler adds handlers for scenarios
func NewScenarioHandler() *ScenarioHandler {
	handler := &ScenarioHandler{}

	// build routes
	var routes []Route
	// Create a Scenario
	routes = append(routes, Route{
		Name:        "Creates a new scenario",
		Method:      http.MethodPost,
		Pattern:     "",
		HandlerFunc: handler.Create,
	})
	// Update a Scenario
	routes = append(routes, Route{
		Name:        "Updates a scenario",
		Method:      http.MethodPut,
		Pattern:     "/{id}",
		HandlerFunc: handler.Update,
	})
	// Get one Scenario
	routes = append(routes, Route{
		Name:        "Get a scenario",
		Method:      http.MethodGet,
		Pattern:     "/{id}",
		HandlerFunc: handler.GetByID,
	})
	// Get all Scenarios
	routes = append(routes, Route{
		Name:        "Get all scenarios",
		Method:      http.MethodGet,
		Pattern:     "",
		HandlerFunc: handler.GetAll,
	})
	// Delete one Scenario
	routes = append(routes, Route{
		Name:        "Delete a scenario",
		Method:      http.MethodDelete,
		Pattern:     "/{id}",
		HandlerFunc: handler.DeleteByID,
	})

	handler.routes = routes

	return handler
}

func (d *ScenarioHandler) Create(w http.ResponseWriter, r *http.Request) {
	body, ok := GetNonEmptyBody(w, r)
	if !ok {
		return
	}

	compiledScenario, err := core.CompileScenario(string(body))
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, exists := repov2.Instance.GetScenario(compiledScenario.GetScenarioDefinition().ID)
	if exists {
		SendJSONError(w, fmt.Sprintf("Scenario with ID %s already exists", compiledScenario.GetScenarioDefinition().ID), http.StatusConflict)
		return
	}

	err = repov2.Instance.SaveScenario((*repov2.PersistedScenario)(compiledScenario.GetScenarioDefinition()))
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (d *ScenarioHandler) Update(w http.ResponseWriter, r *http.Request) {
	body, ok := GetNonEmptyBody(w, r)
	if !ok {
		return
	}

	compiledScenario, err := core.CompileScenario(string(body))
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, exists := repov2.Instance.GetScenario(compiledScenario.GetScenarioDefinition().ID)
	if !exists {
		SendJSONError(w, fmt.Sprintf("Scenario with ID %s does not exist", compiledScenario.GetScenarioDefinition().ID), http.StatusNotFound)
		return
	}

	err = repov2.Instance.SaveScenario((*repov2.PersistedScenario)(compiledScenario.GetScenarioDefinition()))
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (d *ScenarioHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	scenDef, ok := repov2.Instance.GetScenario(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	SendRawStringHTTPCode(w, scenDef.Script, http.StatusOK)
}

type ScenarioMetadata struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (d *ScenarioHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	scenDefs := repov2.Instance.GetAllScenarios()
	scenDefsDTO := make([]ScenarioMetadata, 0, len(scenDefs))
	for _, def := range scenDefs {
		scenDefsDTO = append(scenDefsDTO, ScenarioMetadata{
			ID:   def.ID,
			Name: def.Name,
		})
	}

	SendJSONWithHTTPCode(w, scenDefsDTO, http.StatusOK)
}

func (d *ScenarioHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	ok := repov2.Instance.DeleteScenario(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
