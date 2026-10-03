package main

import (
	"log"
	"net/http"

	"github.com/pickIQ/oredis/internal/httpapi"
	"github.com/pickIQ/oredis/internal/middleware"
	store "github.com/pickIQ/oredis/internal/store"
)

func main() {
	st := store.NewInMemoryStore()
	mux := httpapi.NewRouter(st)

	log.Println("server starting on :8080")
	err := http.ListenAndServe(":8080", middleware.Logging(middleware.Recovery(mux)))
	if err != nil {
		log.Fatal(err)
	}
}
