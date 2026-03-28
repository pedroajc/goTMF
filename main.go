// main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
		Addr:    ":8081",
		Handler: mux,
	}

	log.Printf("Starting server on %s...", server.Addr)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	ctxHttp, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	catalogs, err := fetchRemoteCatalogs(ctxHttp, "http://localhost:8081")
	if err != nil {
		log.Printf("error in startup probe %v", err)
		return
	}
	log.Printf("Startup probe: fetched %d catalogs from local node", len(catalogs))

	select {
	case <-ctx.Done():
		log.Println("Shutting down...")
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown() // ← immediately after creation
		if err := server.Shutdown(ctxShutdown); err != nil {
			log.Printf("shutdown with error: %v", err)
			os.Exit(1)
		}
		log.Println("Server stopped.")
	}
}
