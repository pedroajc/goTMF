// handlers.go
package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"
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
	index, catalog := findCatalog(id)
	if index == -1 {
		writeError(w, http.StatusNotFound, "404", "catalog not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(*catalog); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func handleCreateCatalog(w http.ResponseWriter, r *http.Request) {
	var input Catalog
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "400", "body is not a catalog")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "400", "name is required")
		return
	}

	input.ID = generateID()
	input.Href = fmt.Sprintf("/catalogManagement/v4/catalog/%s", input.ID)
	input.LastUpdate = time.Now().UTC().Format(time.RFC3339)
	if input.AtType == "" {
		input.AtType = "Catalog"
	}

	catalogs = append(catalogs, input)
	go dispatchEvent(buildEvent("CatalogCreateEvent", input))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", input.Href)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(input); err != nil {
		log.Printf("unable to create catalog: %v", err)
	}
}

func handleUpdateCatalog(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := r.PathValue("id")
	index, catalog := findCatalog(id)
	if index == -1 {
		writeError(w, http.StatusNotFound, "404", "catalog not found")
		return
	}

	existing, err := json.Marshal(*catalog)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "500", "internal error")
		return
	}
	var original map[string]any
	if err := json.Unmarshal(existing, &original); err != nil {
		writeError(w, http.StatusInternalServerError, "500", "internal error")
		return
	}

	var overlay map[string]any
	if err := json.NewDecoder(r.Body).Decode(&overlay); err != nil {
		writeError(w, http.StatusBadRequest, "400", "bad request")
		return
	}

	for key, value := range overlay {
		original[key] = value
	}

	newCat, err := json.Marshal(original)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "500", "internal error")
		return
	}
	var updated Catalog
	if err := json.Unmarshal(newCat, &updated); err != nil {
		writeError(w, http.StatusInternalServerError, "500", "internal error")
		return
	}

	updated.ID = catalog.ID
	updated.Href = catalog.Href
	updated.LastUpdate = time.Now().UTC().Format(time.RFC3339)

	catalogs[index] = updated
	go dispatchEvent(buildEvent("CatalogAttributeValueChangeEvent", updated))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updated); err != nil {
		log.Printf("encode error: %v", err)
		return
	}
}

func handleDeleteCatalog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	index, catalog := findCatalog(id)
	if index == -1 {
		writeError(w, http.StatusNotFound, "404", "not found")
		return
	}
	deleted := *catalog

	catalogs = slices.Delete(catalogs, index, index+1)
	go dispatchEvent(buildEvent("CatalogDeleteEvent", deleted))
	w.WriteHeader(http.StatusNoContent)
}

func handleRegisterHub(w http.ResponseWriter, r *http.Request) {
	var body HubSubscription
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "400", "bad request")
		return
	}
	if body.Callback == "" {
		writeError(w, http.StatusBadRequest, "400", "empty callback")
		return
	}

	body.ID = generateID()
	subsMu.Lock()
	subscriptions = append(subscriptions, body)
	subsMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("encode error: %v", err)
	}

}

func handleDeleteHub(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	subsMu.Lock()
	originalLen := len(subscriptions)
	subscriptions = slices.DeleteFunc(subscriptions, func(s HubSubscription) bool { return s.ID == id })
	finalLen := len(subscriptions)
	subsMu.Unlock()

	if originalLen <= finalLen {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func findCatalog(id string) (int, *Catalog) {
	index := slices.IndexFunc(catalogs, func(c Catalog) bool { return c.ID == id })
	if index == -1 {
		return -1, nil
	}
	return index, &catalogs[index]
}
