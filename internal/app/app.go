package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dejaniskra/go-gi/internal/config"
	"github.com/dejaniskra/go-gi/internal/logger"
)

type HTTPMethod string

const (
	GET     HTTPMethod = "GET"
	POST    HTTPMethod = "POST"
	PUT     HTTPMethod = "PUT"
	DELETE  HTTPMethod = "DELETE"
	PATCH   HTTPMethod = "PATCH"
	OPTIONS HTTPMethod = "OPTIONS"
	HEAD    HTTPMethod = "HEAD"
)

type routeKey struct {
	Method string
	Path   string
}

type Application struct {
	routes      map[routeKey]http.HandlerFunc
	middlewares []func(http.Handler) http.Handler
}

func NewApplication() *Application {
	return &Application{
		routes: make(map[routeKey]http.HandlerFunc),
	}
}
func (a *Application) SetLogger(level logger.Level, format logger.Format) {
	logger.InitGlobal(level, format)
}

func (a *Application) AddRoute(method HTTPMethod, path string, handler http.HandlerFunc) {
	a.routes[routeKey{Method: string(method), Path: path}] = handler
}

func (a *Application) Use(mw func(http.Handler) http.Handler) {
	a.middlewares = append(a.middlewares, mw)
}

func (a *Application) Run() error {
	router := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, handler := range a.routes {
			params, matched := matchRoute(key.Path, r.URL.Path)
			if matched && key.Method == r.Method {
				ctx := context.WithValue(r.Context(), "pathParams", params)
				handler(w, r.WithContext(ctx))
				return
			}
		}
		http.NotFound(w, r)
	})

	var handler http.Handler = router
	for i := len(a.middlewares) - 1; i >= 0; i-- {
		handler = a.middlewares[i](handler)
	}

	cfg := config.LoadConfig("config.json")

	srv := &http.Server{
		Addr:              ":" + fmt.Sprintf("%d", *cfg.Http.Port),
		Handler:           handler,
		ReadTimeout:       time.Duration(*cfg.Http.Timeouts.ReadRequest) * time.Second,
		ReadHeaderTimeout: time.Duration(*cfg.Http.Timeouts.ReadRequestHeader) * time.Second,
		WriteTimeout:      time.Duration(*cfg.Http.Timeouts.ResponseWrite) * time.Second,
		IdleTimeout:       time.Duration(*cfg.Http.Timeouts.Idle) * time.Second,
		MaxHeaderBytes:    *cfg.Http.MaxHeaderBytes,
	}

	fmt.Printf("🚀 Server running on %s\n", srv.Addr)
	return srv.ListenAndServe()
}
