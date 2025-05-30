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
	httpServer := GetServer()
	if application.httpServer == nil {
		application.httpServer = httpServer
	} // TODO: revisit this
	httpServer.AddRoute(method, path, handler)
}

func (application *Application) AddMiddleware(mw func(http.Handler) http.Handler) {
	httpServer := GetServer()
	if application.httpServer == nil {
		application.httpServer = httpServer
	} // TODO: revisit this
	httpServer.AddMiddleware(mw)
}

func (application *Application) AddCronJob(name, cronExpr string, fn func()) error {
	return AddJobCron(name, cronExpr, fn)
}

func (application *Application) AddIntervalJob(name string, interval time.Duration, fn func()) error {
	return AddJobInterval(name, interval, fn)
}

func (application *Application) Start() error {
	// ADD cron and kafka here to the chain
	if application.httpServer == nil {
		return fmt.Errorf("no need to start an empty application")
	}

	cfg := config.GetConfig()

	if application.httpServer != nil {
		err := application.httpServer.Start(cfg)
		if err != nil {
			fmt.Println("Error starting HTTP server:", err)
			return err
		}
	}

	return nil
}
