// handlers.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"net/http"
	"slices"
	"strings"
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

	filterParams := r.URL.Query().Get("fields")

	if filterParams == "" {
		if err := json.NewEncoder(w).Encode(catalogs); err != nil {
			log.Printf("encode error: %v", err)
		}
		return
	}

	data, err := json.Marshal(catalogs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "500", "internal error")
		return
	}

	resp, err := filterFields(data, filterParams)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "500", "internal error")
		return
	}

	if _, err := w.Write(resp); err != nil {
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

func filterFields(data []byte, fields string) ([]byte, error) {
	paramFields := strings.Split(fields, ",")

	var filterMap []map[string]any
	if err := json.Unmarshal(data, &filterMap); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	for _, cat := range filterMap {
		maps.DeleteFunc(cat, func(k string, v any) bool { return !slices.Contains(paramFields, k) })
	}

	marMap, err := json.Marshal(filterMap)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}
	return marMap, nil

}
