// federation/main.go

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
	mux.HandleFunc("GET /catalogManagement/v4/catalog", handleFederatedListCatalogs)
	mux.HandleFunc("GET /catalogManagement/v4/catalog/{id}", handleFederatedGetCatalog)

	server := &http.Server{
		Addr:    ":9090",
		Handler: mux,
	}
	log.Printf("Starting server on %s...", server.Addr)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
