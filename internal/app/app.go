package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

	srv := &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Printf("🚀 Server running on %s\n", srv.Addr)
	return srv.ListenAndServe()
}
