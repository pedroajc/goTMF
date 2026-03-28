// federation/main.go

package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalogManagement/v4/catalog", handleFederatedListCatalogs)
	mux.HandleFunc("GET /catalogManagement/v4/catalog/{id}", handleFederatedGetCatalog)

	server := &http.Server{
		Addr:    ":9090",
		Handler: mux,
	}
	log.Printf("Starting server on %s...", server.Addr)
	log.Fatal(server.ListenAndServe())

}
