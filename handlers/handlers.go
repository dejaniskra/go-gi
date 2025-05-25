package handlers

import "github.com/dejaniskra/go-gi/internal/http"

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestHandler(req *http.HTTPRequest, res *http.HTTPResponse) {
	res.Headers["Content-Type"] = "application/json"
	res.Headers["Dejan"] = "Iskra"
	res.StatusCode = 200

	var person Person
	err := req.ToJson(&person)
	if err != nil {
		res.StatusCode = 400
		res.Body = nil
		return
	}
	person.Age += 1
	person.Name += " Iskra"

	reader, err := res.FromJson(person)
	if err != nil {
		res.StatusCode = 500
		res.Body = nil
		return
	}
	res.Body = reader
}

func TestHandlerParam(req *http.HTTPRequest, res *http.HTTPResponse) {
	res.Headers["Content-Type"] = "application/json"
	res.Headers["Dejan"] = req.PathParams["id"]
	res.StatusCode = 404
	res.Body = req.Body
}
