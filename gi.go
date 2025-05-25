package public

import (
	"fmt"

	"github.com/dejaniskra/go-gi/app"
	"github.com/dejaniskra/go-gi/handlers"
	"github.com/dejaniskra/go-gi/logger"
	"github.com/dejaniskra/go-gi/types"
	"github.com/dejaniskra/go-gi/utils"

	"github.com/dejaniskra/go-gi/internal/clients"
	"github.com/dejaniskra/go-gi/internal/middleware"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func GoGiX() *app.Application {
	return app.NewApplication()
}

func xxx() {
	client := clients.NewHTTPClient("http://rocket.com", nil, 10)
	timeout := 5
	person := Person{
		Name: "Dejan",
		Age:  41,
	}

	reqBody, err := utils.JsonToReader(person)
	if err != nil {
		fmt.Println("Error:", err)
	}
	response, err := client.Execute(&clients.HTTPRequest{
		Method:  "POST",
		Path:    "/test",
		Timeout: &timeout,
		Body:    reqBody,
	})
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
	}

	application := app.NewApplication()
	application.SetLogger(logger.INFO, logger.JSON)

	application.AddRoute(types.HTTP_POST, "/testx", handlers.TestHandler)
	application.AddRoute(types.HTTP_POST, "/testx/:id", handlers.TestHandlerParam)

	application.AddMiddleware(middleware.RecoverMiddleware)
	application.AddMiddleware(middleware.RequestIDMiddleware)

	err = application.Start()

	if err != nil {
		panic("failed to start application: " + err.Error())
	}
}
