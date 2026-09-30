package main

import (
	"log"
	"net/http"

	"github.com/pickIQ/oredis/internal/httpapi"
	store "github.com/pickIQ/oredis/internal/store"
)

func main() {
	st := store.NewInMemoryStore()
	mux := httpapi.NewRouter(st)

	log.Println("server starting on :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
