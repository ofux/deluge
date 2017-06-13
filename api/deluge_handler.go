package api

// DelugeHandler handles requests for 'deluge' resource
type DelugeHandler struct {
	routes []Route
}

func (d *DelugeHandler) GetBasePath() string {
	return "/deluges"
}

func (d *DelugeHandler) GetRoutes() []Route {
	return d.routes
}

// NewTaskController creates a new task controller to manage tasks
func NewDelugeHandler() *DelugeHandler {
	controller := &DelugeHandler{}

	// build routes
	routes := []Route{}
	// GetAll
	/*routes = append(routes, Route{
		Name:        "Get all tasks",
		Method:      http.MethodGet,
		Pattern:     "",
		HandlerFunc: controller.GetAll,
	})
	// Get
	routes = append(routes, Route{
		Name:        "Get one task",
		Method:      http.MethodGet,
		Pattern:     "/{id}",
		HandlerFunc: controller.Get,
	})
	// Create
	routes = append(routes, Route{
		Name:        "Create a task",
		Method:      http.MethodPost,
		Pattern:     "",
		HandlerFunc: controller.Create,
	})
	// Update
	routes = append(routes, Route{
		Name:        "Update a task",
		Method:      http.MethodPut,
		Pattern:     "/{id}",
		HandlerFunc: controller.Update,
	})
	// Delete
	routes = append(routes, Route{
		Name:        "Delete a task",
		Method:      http.MethodDelete,
		Pattern:     "/{id}",
		HandlerFunc: controller.Delete,
	})*/

	controller.routes = routes

	return controller
}
