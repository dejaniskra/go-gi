package gogi

import (
	"fmt"
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
	Path        string
	Headers     map[string]string
	Body        io.Reader
	QueryParams map[string]string
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
	baseUrl := ""

	if c.BaseURL != nil {
		baseUrl = strings.TrimRight(*c.BaseURL, "/")
	}

	fullURL := baseUrl + "/" + strings.TrimLeft(req.Path, "/")

	httpReq, err := http.NewRequest(req.Method, fullURL, req.Body)

	if err != nil {
		return nil, err
	}

	if req.Headers != nil {
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}

	timeout := c.Timeout
	if req.Timeout != nil {
		timeout = req.Timeout
	}

	client := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}
	defer resp.Body.Close()

	response := &HTTPClientResponse{
		StatusCode: resp.StatusCode,
		Headers:    make(map[string]string),
		Body:       resp.Body,
	}

	for key, value := range resp.Header {
		response.Headers[key] = value[0]
	}

	return response, nil
}

func (r *HTTPRequest) ToJSON(dest interface{}) error {
	return ReaderToJson(r.Body, dest)
}

func (r *HTTPResponse) FromJSON(v interface{}) (io.Reader, error) {
	return JsonToReader(v)
}
