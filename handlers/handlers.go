package handlers

import (
	"net/http"

	"github.com/dejaniskra/go-gi/internal/app"
)

func HelloHandler(a *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.JSON(w, map[string]string{"hello": "world"}, http.StatusOK)
	}
}

func GetUserHandler(a *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := a.Param(r, "id")
		a.JSON(w, map[string]string{"user_id": id}, http.StatusOK)
	}
}
