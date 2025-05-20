package main

import (
	"github.com/dejaniskra/go-gi/handlers"
	"github.com/dejaniskra/go-gi/internal/app"
)

func main() {
	a := app.NewApplication()

	// application.SetLogger(logger.INFO, logger.JSON)
	// application.Use(middleware.RecoverMiddleware)
	// application.Use(middleware.RequestIDMiddleware)

	a.Get("/hello", handlers.HelloHandler(a))
	a.Get("/hello/:id", handlers.GetUserHandler(a))

	err := a.Run()

	if err != nil {
		panic("failed to start server: " + err.Error())
	}
}
