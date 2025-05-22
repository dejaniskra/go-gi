package main

import (
	"github.com/dejaniskra/go-gi/handlers"
	"github.com/dejaniskra/go-gi/internal/app"
	"github.com/dejaniskra/go-gi/internal/http"
	"github.com/dejaniskra/go-gi/internal/middleware"
	"github.com/dejaniskra/go-gi/pkg/shared/logger"
)

func main() {
	application := app.NewApplication()
	application.SetLogger(logger.INFO, logger.JSON)

	server := application.NewHttpServer()

	server.AddRoute(http.GET, "/hello", handlers.Hello)
	server.AddRoute(http.GET, "/hello/:id", handlers.GetUser)

	server.AddMiddleware(middleware.RecoverMiddleware)
	server.AddMiddleware(middleware.RequestIDMiddleware)

	err := application.Start()

	if err != nil {
		panic("failed to start application: " + err.Error())
	}
}
