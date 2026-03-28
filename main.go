// main.go
package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalogManagement/v4/catalog", handleListCatalogs)
	mux.HandleFunc("GET /catalogManagement/v4/catalog/{id}", handleGetCatalog)
	mux.HandleFunc("POST /catalogManagement/v4/catalog", handleCreateCatalog)
	mux.HandleFunc("PATCH /catalogManagement/v4/catalog/{id}", handleUpdateCatalog)
	mux.HandleFunc("DELETE /catalogManagement/v4/catalog/{id}", handleDeleteCatalog)
	mux.HandleFunc("POST /catalogManagement/v4/hub", handleRegisterHub)
	mux.HandleFunc("DELETE /catalogManagement/v4/hub/{id}", handleDeleteHub)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Starting server on %s...", server.Addr)
	go func() {
		log.Fatal(server.ListenAndServe())
	}()
	// give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	catalogs, err := fetchRemoteCatalogs(ctx, "http://localhost:8080")
	if err != nil {
		log.Printf("error in startup probe %v", err)
		return
	}
	log.Printf("Startup probe: fetched %d catalogs from local node", len(catalogs))
	select {}

}
