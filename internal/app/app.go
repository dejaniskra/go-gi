package app

import (
	"fmt"

	"github.com/dejaniskra/go-gi/internal/config"
	"github.com/dejaniskra/go-gi/internal/http"
	"github.com/dejaniskra/go-gi/pkg/shared/logger"
)

type Application struct {
	HttpServer *http.HttpServer
}

func NewApplication() *Application {
	return &Application{}
}

func (application *Application) SetLogger(level logger.Level, format logger.Format) {
	logger.InitGlobal(level, format)
}

func (application *Application) NewHttpServer() *http.HttpServer {
	httpServer := http.NewServer()

	application.HttpServer = httpServer
	return application.HttpServer
}

func (application *Application) Start() error {
	// ADD cron and kafka here to the chain
	if application.HttpServer == nil {
		return fmt.Errorf("no need to start an empty application")
	}

	cfg := config.LoadConfig("config.json")

	if application.HttpServer != nil {
		err := application.HttpServer.Start(cfg)
		if err != nil {
			fmt.Println("Error starting HTTP server:", err)
			return err
		}
	}

	return nil
}
