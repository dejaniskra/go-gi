package http

import (
	"fmt"
	"io"
	"net/http"
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

func Handler(handler HTTPHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &HTTPRequest{
			Method:      r.Method,
			PathParams:  make(map[string]string),
			QueryParams: make(map[string]string),
			Headers:     make(map[string]string),
			Body:        r.Body,
		}
		for k, v := range r.Header {
			req.Headers[k] = v[0] // Simplified: only first header value; TODO: // verify
		}
		for k, v := range r.URL.Query() {
			req.QueryParams[k] = v[0] // Simplified: only first query value; TODO: // verify
		}
		params, ok := r.Context().Value("pathParams").(map[string]string)
		if ok {
			for k, v := range params {
				req.PathParams[k] = v
			}
		}

		fmt.Println("QueryParams:", req.QueryParams)
		fmt.Println("PathParams:", req.PathParams)

		res := &HTTPResponse{
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
