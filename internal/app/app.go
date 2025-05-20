package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const pathParamsKey = contextKey("pathParams")

type Application struct {
	routes      []route
	middlewares []func(http.Handler) http.Handler
}

type route struct {
	method  string
	pattern string
	handler http.HandlerFunc
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Get(pattern string, handler http.HandlerFunc) {
	app.routes = append(app.routes, route{
		method:  "GET",
		pattern: pattern,
		handler: handler,
	})
}
func (app *Application) Post(pattern string, handler http.HandlerFunc) {
	app.routes = append(app.routes, route{
		method:  "POST",
		pattern: pattern,
		handler: handler,
	})
}

func (app *Application) Patch(pattern string, handler http.HandlerFunc) {
	app.routes = append(app.routes, route{
		method:  "PATCH",
		pattern: pattern,
		handler: handler,
	})
}

func (app *Application) Put(pattern string, handler http.HandlerFunc) {
	app.routes = append(app.routes, route{
		method:  "PUT",
		pattern: pattern,
		handler: handler,
	})
}

func (app *Application) Delete(pattern string, handler http.HandlerFunc) {
	app.routes = append(app.routes, route{
		method:  "DELETE",
		pattern: pattern,
		handler: handler,
	})
}

func (app *Application) Use(mw func(http.Handler) http.Handler) {
	app.middlewares = append(app.middlewares, mw)
}

func (app *Application) Run() error {
	router := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler, params := app.matchRoute(r)
		if handler != nil {
			ctx := context.WithValue(r.Context(), pathParamsKey, params)
			handler(w, r.WithContext(ctx))
			return
		}
		http.NotFound(w, r)
	})

	var handler http.Handler = router
	for i := len(app.middlewares) - 1; i >= 0; i-- {
		handler = app.middlewares[i](handler)
	}

	srv := &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	println("🚀 Server running on http://localhost:8080")
	return srv.ListenAndServe()
}

// -- Path parameter logic --

func (app *Application) matchRoute(r *http.Request) (http.HandlerFunc, map[string]string) {
	for _, rt := range app.routes {
		if r.Method != rt.method {
			continue
		}
		params, ok := matchPath(rt.pattern, r.URL.Path)
		if ok {
			return rt.handler, params
		}
	}
	return nil, nil
}

func matchPath(pattern, path string) (map[string]string, bool) {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i := 0; i < len(patternParts); i++ {
		if strings.HasPrefix(patternParts[i], ":") {
			params[patternParts[i][1:]] = pathParts[i]
		} else if patternParts[i] != pathParts[i] {
			return nil, false
		}
	}
	return params, true
}

// -- Helpers --

func (app *Application) Param(r *http.Request, key string) string {
	params, ok := r.Context().Value(pathParamsKey).(map[string]string)
	if !ok {
		return ""
	}
	return params[key]
}

func (app *Application) JSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
