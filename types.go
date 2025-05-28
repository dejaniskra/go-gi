package gogi

import (
	"io"
	"net/http"
)

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
