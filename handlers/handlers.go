package handlers

import "github.com/dejaniskra/go-gi/internal/http"

func TestHandler(req *http.HTTPRequest, res *http.HTTPResponse) {
	res.Headers["Content-Type"] = "application/json"
	res.Headers["Dejan"] = "Iskra"
	res.StatusCode = 200
	res.Body = req.Body
}

func TestHandlerParam(req *http.HTTPRequest, res *http.HTTPResponse) {
	res.Headers["Content-Type"] = "application/json"
	res.Headers["Dejan"] = req.PathParams["id"]
	res.StatusCode = 404
	res.Body = req.Body
}
