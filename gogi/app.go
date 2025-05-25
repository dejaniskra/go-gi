package app

import (
	"fmt"
	"net/http"

	"github.com/dejaniskra/go-gi/internal/config"
	gogiHttp "github.com/dejaniskra/go-gi/internal/http"
	"github.com/dejaniskra/go-gi/logger"
	"github.com/dejaniskra/go-gi/types"
)

type Application struct {
	httpServer *gogiHttp.HttpServer
}

func NewApplication() *Application {
	return &Application{}
}

func (application *Application) SetLogger(level logger.Level, format logger.Format) {
	logger.InitGlobal(level, format)
}

func (application *Application) AddRoute(method types.HTTPMethod, path string, handler gogiHttp.HTTPHandler) {
	httpServer := gogiHttp.GetServer()
	if application.httpServer == nil {
		application.httpServer = httpServer
	} // TODO: revisit this
	httpServer.AddRoute(method, path, handler)
}

func (application *Application) AddMiddleware(mw func(http.Handler) http.Handler) {
	httpServer := gogiHttp.GetServer()
	if application.httpServer == nil {
		application.httpServer = httpServer
	} // TODO: revisit this
	httpServer.AddMiddleware(mw)
}

func (application *Application) Start() error {
	// ADD cron and kafka here to the chain
	if application.httpServer == nil {
		return fmt.Errorf("no need to start an empty application")
	}

	cfg := config.LoadConfig("config.json")

	if application.httpServer != nil {
		err := application.httpServer.Start(cfg)
		if err != nil {
			fmt.Println("Error starting HTTP server:", err)
			return err
		}
	}

	return nil
}
