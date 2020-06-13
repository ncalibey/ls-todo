package main

import (
	"net/http"

	"github.com/gorilla/mux"

	"ls-todo/internal/server"
)

func main() {
	router := mux.NewRouter()
	s := server.New(router)

	http.ListenAndServe(":8080", s)
}
