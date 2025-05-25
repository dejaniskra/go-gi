package http

import (
	"io"
	"net/http"
)

type routeKey struct {
	Method string
	Path   string
}

type HttpServer struct {
	middlewares []func(http.Handler) http.Handler
	routes      map[routeKey]http.HandlerFunc
}

type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       io.Reader
}

type HTTPRequest struct {
	Method      string
	PathParams  map[string]string
	QueryParams map[string]string
	Headers     map[string]string
	Body        io.Reader
}

type HTTPHandler func(*HTTPRequest, *HTTPResponse)
type MiddlewareHandler func(http.Handler) http.Handler
