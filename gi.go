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

	application.AddRoute(http.POST, "/testx", handlers.TestHandler)
	application.AddRoute(http.POST, "/testx/:id", handlers.TestHandlerParam)

	application.AddMiddleware(middleware.RecoverMiddleware)
	application.AddMiddleware(middleware.RequestIDMiddleware)

	err := application.Start()

	if err != nil {
		panic("failed to start application: " + err.Error())
	}
}
