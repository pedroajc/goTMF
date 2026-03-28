// federation/gateway.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

var nodes = []NodeConfig{
	{ID: "node-a", BaseURL: "http://localhost:8081", Name: "Node A"},
	{ID: "node-b", BaseURL: "http://localhost:8082", Name: "Node B"},
}

func federatedListCatalogs(ctx context.Context) []Catalog {
	receivedCatalogs := make(chan ([]Catalog), len(nodes))
	var wg sync.WaitGroup

	for _, node := range nodes {
		wg.Add(1)
		go func(n NodeConfig) {
			defer wg.Done()
			catalog, err := fetchRemoteCatalogs(ctx, n.BaseURL)
			if err != nil {
				log.Printf("query to node %s failed: %v", n.ID, err)
				return
			}
			receivedCatalogs <- catalog
		}(node)
	}

	go func() {
		wg.Wait()
		close(receivedCatalogs)
	}()

	var results []Catalog
	for r := range receivedCatalogs {
		results = append(results, r...)
	}

	return deduplicate(results)
}

func deduplicate(source []Catalog) []Catalog {
	deduplicated := make([]Catalog, 0)
	seen := make(map[string]bool)
	for _, c := range source {
		if seen[c.ID] {
			continue
		}
		seen[c.ID] = true
		deduplicated = append(deduplicated, c)
	}
	return deduplicated
}

func federatedGetCatalog(ctx context.Context, id string) *Catalog {
	receivedResponses := make(chan *Catalog, len(nodes))
	var wg sync.WaitGroup

	for _, node := range nodes {
		wg.Add(1)
		go func(n NodeConfig) {
			defer wg.Done()
			catalog, err := fetchRemoteCatalog(ctx, n.BaseURL, id)
			if err != nil {
				log.Printf("query to node %s failed: %v", n.ID, err)
				return
			}
			if catalog != nil {
				receivedResponses <- catalog
			}
		}(node)
	}

	go func() {
		wg.Wait()
		close(receivedResponses)
	}()

	for catalog := range receivedResponses {
		return catalog
	}
	return nil
}

func handleFederatedListCatalogs(w http.ResponseWriter, r *http.Request) {
	catalogs := federatedListCatalogs(r.Context())
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(catalogs); err != nil {
		log.Printf("internal server error: %v", err)
		return
	}

}

func handleFederatedGetCatalog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	catalog := federatedGetCatalog(r.Context(), id)
	if catalog == nil {
		writeError(w, http.StatusNotFound, "404", "catalog not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(catalog); err != nil {
		log.Printf("internal server error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Error{Code: code, Reason: reason})
}
