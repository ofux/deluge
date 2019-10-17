package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/repov2"
	"net/http"
)

// DelugeHandler handles requests for 'deluges' resource
type DelugeHandler struct {
	routes []Route
}

func (d *DelugeHandler) GetBasePath() string {
	return "/v1/deluges"
}

func (d *DelugeHandler) GetRoutes() []Route {
	return d.routes
}

// NewDelugeHandler adds handlers for deluges
func NewDelugeHandler() *DelugeHandler {
	handler := &DelugeHandler{}

	// build routes
	var routes []Route
	// Create a Deluge
	routes = append(routes, Route{
		Name:        "Creates a new deluge",
		Method:      http.MethodPost,
		Pattern:     "",
		HandlerFunc: handler.Create,
	})
	// Update a Deluge
	routes = append(routes, Route{
		Name:        "Updates a deluge",
		Method:      http.MethodPut,
		Pattern:     "/{id}",
		HandlerFunc: handler.Update,
	})
	// Get one Deluge
	routes = append(routes, Route{
		Name:        "Get a deluge",
		Method:      http.MethodGet,
		Pattern:     "/{id}",
		HandlerFunc: handler.GetByID,
	})
	// Get all Deluges
	routes = append(routes, Route{
		Name:        "Get all deluges",
		Method:      http.MethodGet,
		Pattern:     "",
		HandlerFunc: handler.GetAll,
	})
	// Delete one Deluge
	routes = append(routes, Route{
		Name:        "Delete a deluge",
		Method:      http.MethodDelete,
		Pattern:     "/{id}",
		HandlerFunc: handler.DeleteByID,
	})

	handler.routes = routes

	return handler
}

func (d *DelugeHandler) Create(w http.ResponseWriter, r *http.Request) {
	body, ok := GetNonEmptyBody(w, r)
	if !ok {
		return
	}

	compiledDeluge, err := core.CompileDeluge(string(body))
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, exists := repov2.Instance.GetDeluge(compiledDeluge.GetDelugeDefinition().ID)
	if exists {
		SendJSONError(w, fmt.Sprintf("Deluge with ID %s already exists", compiledDeluge.GetDelugeDefinition().ID), http.StatusConflict)
		return
	}

	err = repov2.Instance.SaveDeluge(compiledDeluge.MapToPersistedDeluge())
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (d *DelugeHandler) Update(w http.ResponseWriter, r *http.Request) {
	body, ok := GetNonEmptyBody(w, r)
	if !ok {
		return
	}

	compiledDeluge, err := core.CompileDeluge(string(body))
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, exists := repov2.Instance.GetDeluge(compiledDeluge.GetDelugeDefinition().ID)
	if !exists {
		SendJSONError(w, fmt.Sprintf("Deluge with ID %s does not exist", compiledDeluge.GetDelugeDefinition().ID), http.StatusNotFound)
		return
	}

	err = repov2.Instance.SaveDeluge(compiledDeluge.MapToPersistedDeluge())
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (d *DelugeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	delugeDef, ok := repov2.Instance.GetDeluge(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	SendRawStringHTTPCode(w, delugeDef.Script, http.StatusOK)
}

type DelugeMetadata struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (d *DelugeHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	delugeDef := repov2.Instance.GetAllDeluges()
	delugeDefsDTO := make([]DelugeMetadata, 0, len(delugeDef))
	for _, def := range delugeDef {
		delugeDefsDTO = append(delugeDefsDTO, DelugeMetadata{
			ID:   def.ID,
			Name: def.Name,
		})
	}

	SendJSONWithHTTPCode(w, delugeDefsDTO, http.StatusOK)
}

func (d *DelugeHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	ok := repov2.Instance.DeleteDeluge(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
