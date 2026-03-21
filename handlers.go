// handlers.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
)

var catalogs = []Catalog{
	{ID: "cat-001", Href: "/catalogManagement/v4/catalog/cat-001",
		Name: "B2B Catalogue", LifecycleStatus: "Active", AtType: "Catalog"},
	{ID: "cat-002", Href: "/catalogManagement/v4/catalog/cat-002",
		Name: "Retail Catalogue", LifecycleStatus: "Active", AtType: "Catalog"},
}

func writeError(w http.ResponseWriter, status int, code, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Error{Code: code, Reason: reason})
}

func handleListCatalogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(catalogs); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func handleGetCatalog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	index := slices.IndexFunc(catalogs, func(c Catalog) bool { return c.ID == id })
	if index == -1 {
		writeError(w, http.StatusNotFound, "404", "catalog not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(catalogs[index]); err != nil {
		log.Printf("encode error: %v", err)
	}
}
