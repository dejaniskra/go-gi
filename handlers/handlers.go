package handlers

import (
	"fmt"
	"net/http"

	internalhttp "github.com/dejaniskra/go-gi/internal/http"
)

func Hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"hello": "world"}`))
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	id := internalhttp.Param(r, "id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"user_id": "%s"}`, id)))
}
