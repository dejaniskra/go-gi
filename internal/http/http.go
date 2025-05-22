package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dejaniskra/go-gi/internal/config"
)

type RouteKey struct {
	Method string
	Path   string
}

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

type HttpServer struct {
	routes      map[RouteKey]http.HandlerFunc
	middlewares []func(http.Handler) http.Handler
}

func NewServer() *HttpServer {
	return &HttpServer{
		routes: make(map[RouteKey]http.HandlerFunc),
	}
}

func (httpServer *HttpServer) AddRoute(method HTTPMethod, path string, handler http.HandlerFunc) {
	httpServer.routes[RouteKey{Method: string(method), Path: path}] = handler
}

func (httpServer *HttpServer) AddMiddleware(mw func(http.Handler) http.Handler) {
	httpServer.middlewares = append(httpServer.middlewares, mw)
}

func (httpServer *HttpServer) Start(cfg *config.Config) error {
	router := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, handler := range httpServer.routes {
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
	for i := len(httpServer.middlewares) - 1; i >= 0; i-- {
		handler = httpServer.middlewares[i](handler)
	}

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
