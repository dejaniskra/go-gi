package gogi

import (
	"github.com/dejaniskra/go-gi/handlers"
	"github.com/dejaniskra/go-gi/internal/app"
	"github.com/dejaniskra/go-gi/internal/clients"
	"github.com/dejaniskra/go-gi/internal/middleware"

	"github.com/dejaniskra/go-gi/pkg/shared/logger"
	"github.com/dejaniskra/go-gi/pkg/shared/types"
)

func NewGoGi() *app.Application {
	return app.NewApplication()
}

func NewAPIClient(baseURL string, headers map[string]string, timeout int) *clients.HTTPClient {
	return clients.NewHTTPClient(baseURL, headers, timeout)
}

func Logger() *logger.Logger {
	return logger.GetLogger()
}

func xxx() {
	application := app.NewApplication()

	application.AddRoute(types.HTTP_POST, "/testx", handlers.TestHandler)
	application.AddRoute(types.HTTP_POST, "/testx/:id", handlers.TestHandlerParam)

	application.AddMiddleware(middleware.RecoverMiddleware)
	application.AddMiddleware(middleware.RequestIDMiddleware)

	err := application.Start()

	if err != nil {
		panic("failed to start application: " + err.Error())
	}
}
