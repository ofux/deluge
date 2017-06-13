package api

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type ResourceHandler interface {
	GetBasePath() string
	GetRoutes() []Route
}

// Router is the struct use for routing
type Router struct {
	*mux.Router
}

// Route is a structure of Route
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// NewRouter creates a new router instance
func NewRouter(resourcesHandlers ...ResourceHandler) *Router {
	router := &Router{mux.NewRouter()}
	router.StrictSlash(false)

	// add routes of handlers
	for _, resourceHandler := range resourcesHandlers {
		for _, route := range resourceHandler.GetRoutes() {
			log.WithField("route", route).Debug("adding route to mux")
			router.
				Methods(route.Method).
				Path(resourceHandler.GetBasePath() + route.Pattern).
				Name(route.Name).
				Handler(route.HandlerFunc)
		}
	}

	return router
}
