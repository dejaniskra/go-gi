package main

import (
	"github.com/dejaniskra/go-gi/handlers"
	"github.com/dejaniskra/go-gi/internal/app"
	"github.com/dejaniskra/go-gi/internal/logger"
	"github.com/dejaniskra/go-gi/internal/middleware"
)

func main() {
	application := app.NewApplication()

	application.SetLogger(logger.INFO, logger.JSON)

	application.Use(middleware.RecoverMiddleware)
	application.Use(middleware.RequestIDMiddleware)

	application.AddRoute(app.GET, "/hello", handlers.Hello)
	application.AddRoute(app.GET, "/hello/:id", handlers.GetUser)

	err := application.Run()

	if err != nil {
		panic("failed to start server: " + err.Error())
	}
}
