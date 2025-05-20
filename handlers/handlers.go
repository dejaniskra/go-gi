package handlers

import (
	"fmt"
	"net/http"

	"github.com/dejaniskra/go-gi/internal/app"
)

func Hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"hello": "world"}`))
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	id := app.Param(r, "id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"user_id": "%s"}`, id)))
}
