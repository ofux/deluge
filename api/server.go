package api

import (
	"github.com/meatballhat/negroni-logrus"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"strconv"
)

func NewServer() *negroni.Negroni {
	// web server
	n := negroni.New()

	// add middleware for logging
	n.Use(negronilogrus.NewMiddlewareFromLogger(log.StandardLogger(), "task"))

	// add recovery middleware in case of panic in handler func
	recovery := negroni.NewRecovery()
	recovery.PrintStack = false
	n.Use(recovery)

	// route handler goes last
	n.UseHandler(NewRouter(NewDelugeHandler()))

	return n
}

func Serve(port int) {
	NewServer().Run(":" + strconv.Itoa(port))
}
