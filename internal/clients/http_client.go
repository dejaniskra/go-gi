package clients

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dejaniskra/go-gi/utils"
)

type HTTPClient struct {
	BaseURL string
	Headers map[string]string
	Timeout int
}
type HTTPRequest struct {
	Method      string
	Path        string
	Headers     map[string]string
	Body        io.Reader
	QueryParams map[string]string
	Timeout     *int
}
type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       io.Reader
}

func NewHTTPClient(baseURL string, headers map[string]string, timeout int) *HTTPClient {
	return &HTTPClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Headers: headers,
		Timeout: timeout,
	}
}
func (c *HTTPClient) Execute(req *HTTPRequest) (*HTTPResponse, error) {
	fullURL := c.BaseURL + "/" + strings.TrimLeft(req.Path, "/")

	httpReq, err := http.NewRequest(req.Method, fullURL, req.Body)

	if err != nil {
		return nil, err
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	timeout := c.Timeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}
	defer resp.Body.Close()

	response := &HTTPResponse{
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
	return utils.ReaderToJson(r.Body, dest)
}

func (r *HTTPResponse) FromJson(v interface{}) (io.Reader, error) {
	return utils.JsonToReader(v)
}
