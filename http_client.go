package gogi

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPClient struct {
	BaseURL *string
	Headers *map[string]string
	Timeout *int
}

type HTTPClientRequest struct {
	Method      string
	Path        *string
	Headers     *map[string]string
	Body        *io.Reader
	QueryParams *map[string]string
	Timeout     *int
}

type HTTPClientResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       io.Reader
}

func NewHTTPClient(baseURL *string, headers *map[string]string, timeout *int) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Headers: headers,
		Timeout: timeout,
	}
}
func (c *HTTPClient) Execute(req *HTTPClientRequest) (*HTTPClientResponse, error) {
	var baseUrl, path string

	if c.BaseURL != nil {
		baseUrl = strings.TrimRight(*c.BaseURL, "/")
	}
	if req.Path != nil {
		path = strings.TrimLeft(*req.Path, "/")
	}

	fullURL := baseUrl
	if path != "" {
		if baseUrl != "" {
			fullURL += "/" + path
		} else {
			fullURL = "/" + path
		}
	}

	if fullURL == "" {
		fullURL = "/"
	}

	var body io.Reader
	if req.Body != nil {
		body = *req.Body
	}

	httpReq, err := http.NewRequest(req.Method, fullURL, body)

	if err != nil {
		return nil, err
	}

	if req.Headers != nil {
		for key, value := range *req.Headers {
			httpReq.Header.Set(key, value)
		}
	}

	if c.Headers != nil {
		for key, value := range *c.Headers {
			if _, exists := httpReq.Header[key]; !exists {
				httpReq.Header.Set(key, value)
			}
		}
	}

	timeout := 30
	if c.Timeout != nil {
		timeout = *c.Timeout
	}
	if req.Timeout != nil {
		timeout = *req.Timeout
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	response := &HTTPClientResponse{
		StatusCode: resp.StatusCode,
		Headers:    make(map[string]string),
		Body:       bytes.NewReader(data),
	}

	for key, values := range resp.Header {
		if len(values) > 0 {
			response.Headers[key] = values[0]
		}
	}

	return response, nil
}
