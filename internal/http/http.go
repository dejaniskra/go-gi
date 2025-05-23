package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dejaniskra/go-gi/internal/config"
)

var httpServer *HttpServer

func GetServer() *HttpServer {
	if httpServer == nil {
		httpServer = &HttpServer{
			routes: make(map[routeKey]http.HandlerFunc),
		}
	}

	return httpServer
}

func (httpServer *HttpServer) AddRoute(method HTTPMethod, path string, handler HTTPHandler) {
	httpServer.routes[routeKey{Method: string(method), Path: path}] = Handler(handler)
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
