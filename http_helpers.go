package gogi

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func httpHandler(handler HTTPHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Context().Value("transaction_id"))
		req := &HTTPRequest{
			Method:      r.Method,
			PathParams:  make(map[string]string),
			QueryParams: make(map[string]string),
			Headers:     make(map[string]string),
			Body:        r.Body,
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
