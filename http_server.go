package gogi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	httpServer.routes[routeKey{Method: string(method), Path: path}] = httpHandler(handler)
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

func httpHandler(handler HTTPHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &HTTPServerRequest{
			Method:      r.Method,
			Path:        r.URL.Path,
			PathParams:  make(map[string]string),
			QueryParams: make(map[string]string),
			Headers:     make(map[string]string),
			Body:        r.Body,
			Context:     r.Context(),
		}
		for k, v := range r.Header {
			req.Headers[k] = v[0]
		}
		for k, v := range r.URL.Query() {
			req.QueryParams[k] = v[0]
		}
		params, ok := r.Context().Value("pathParams").(map[string]string)
		if ok {
			for k, v := range params {
				req.PathParams[k] = v
			}
		}

		res := &HTTPServerResponse{
			Headers: make(map[string]string),
		}

		handler(req, res)

		if res.StatusCode == 0 {
			res.StatusCode = http.StatusOK // Default
		}

		for k, v := range res.Headers {
			w.Header().Set(k, v)
		}

		w.WriteHeader(res.StatusCode)

		if res.Body != nil {
			io.Copy(w, res.Body)
		}
	}
}

func matchRoute(pattern, actual string) (map[string]string, bool) {
	pParts := strings.Split(strings.Trim(pattern, "/"), "/")
	aParts := strings.Split(strings.Trim(actual, "/"), "/")

	if len(pParts) != len(aParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i := range pParts {
		if strings.HasPrefix(pParts[i], ":") {
			params[strings.TrimPrefix(pParts[i], ":")] = aParts[i]
		} else if pParts[i] != aParts[i] {
			return nil, false
		}
	}

	return params, true
}

type routeKey struct {
	Method string
	Path   string
}

type HttpServer struct {
	middlewares []func(http.Handler) http.Handler
	routes      map[routeKey]http.HandlerFunc
}

type HTTPServerResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       io.Reader
}

type HTTPServerRequest struct {
	Method      string
	Path        string
	PathParams  map[string]string
	QueryParams map[string]string
	Headers     map[string]string
	Body        io.Reader
	Context     context.Context
}

type HTTPHandler func(*HTTPServerRequest, *HTTPServerResponse)
type MiddlewareHandler func(http.Handler) http.Handler

type HTTPMethod string

const (
	HTTP_GET     HTTPMethod = "GET"
	HTTP_POST    HTTPMethod = "POST"
	HTTP_PUT     HTTPMethod = "PUT"
	HTTP_DELETE  HTTPMethod = "DELETE"
	HTTP_PATCH   HTTPMethod = "PATCH"
	HTTP_OPTIONS HTTPMethod = "OPTIONS"
	HTTP_HEAD    HTTPMethod = "HEAD"
)
