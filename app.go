package gogi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dejaniskra/go-gi/internal/config"
)

type Application struct {
	httpServer *HttpServer
}

func NewApplication() *Application {
	return &Application{}
}

func (application *Application) AddRoute(method HTTPMethod, path string, handler HTTPHandler) {
	httpServer := getServer()
	if application.httpServer == nil {
		application.httpServer = httpServer
	} // TODO: revisit this
	httpServer.addRoute(method, path, handler)
}

func (application *Application) AddMiddleware(mw func(http.Handler) http.Handler) {
	httpServer := getServer()
	if application.httpServer == nil {
		application.httpServer = httpServer
	} // TODO: revisit this
	httpServer.addMiddleware(mw)
}

func (application *Application) AddCronJob(name, cronExpr string, fn func()) error {
	return addJobCron(name, cronExpr, fn)
}

func (application *Application) AddIntervalJob(name string, interval time.Duration, fn func()) error {
	return addJobInterval(name, interval, fn)
}

func (application *Application) Start() error {
	if application.httpServer == nil {
		return fmt.Errorf("no need to start an empty application")
	}

	cfg := config.GetConfig()

	if application.httpServer != nil {
		err := application.httpServer.start(cfg)
		if err != nil {
			fmt.Println("Error starting HTTP server:", err)
			return err
		}
	}

	return nil
}
