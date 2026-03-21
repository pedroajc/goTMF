// main.go
package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalogManagement/v4/catalog", handleListCatalogs)
	mux.HandleFunc("GET /catalogManagement/v4/catalog/{id}", handleGetCatalog)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Printf("Starting server on %s...\n", server.Addr)
	log.Fatal(server.ListenAndServe())

}
