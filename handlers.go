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
	"strconv"
	"strings"
	"sync"
	"time"
)

var catalogs = []Catalog{
	{ID: "cat-001", Href: "/catalogManagement/v4/catalog/cat-001",
		Name: "B2B Catalogue", LifecycleStatus: "Active", AtType: "Catalog"},
	{ID: "cat-002", Href: "/catalogManagement/v4/catalog/cat-002",
		Name: "Retail Catalogue", LifecycleStatus: "Active", AtType: "Catalog"},
}

var muCat sync.RWMutex

func writeError(w http.ResponseWriter, status int, code, reason string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Error{Code: code, Reason: reason, Message: message, Status: strconv.Itoa(status)})
}

func handleListCatalogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	filterParams := r.URL.Query().Get("fields")
	muCat.RLock()
	data, err := json.Marshal(catalogs)
	muCat.RUnlock()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "500", "internal error", "unable to marshal catalogs")
		return
	}

	if filterParams == "" {
		if _, err := w.Write(data); err != nil {
			log.Printf("encode error: %v", err)
		}
		return
	}

	resp, err := filterFields(data, filterParams)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "500", "internal error", "unable to filter the catalogs")
		return
	}

	if _, err := w.Write(resp); err != nil {
		log.Printf("encode error: %v", err)
	}

}

func handleGetCatalog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	muCat.RLock()
	index, foundCat := findCatalog(id)
	if index == -1 {
		muCat.RUnlock()
		writeError(w, http.StatusNotFound, "404", "catalog not found", "ID doesn't exist in the store")
		return
	}
	catalog := *foundCat
	muCat.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(catalog); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func handleCreateCatalog(w http.ResponseWriter, r *http.Request) {
	var input Catalog
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "400", "malformed input", "body is not a catalog")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "400", "malformed input", "name is required")
		return
	}

	input.LastUpdate = time.Now().UTC().Format(time.RFC3339)
	if input.AtType == "" {
		input.AtType = "Catalog"
	}
	muCat.Lock()
	input.ID = generateID()
	input.Href = fmt.Sprintf("/catalogManagement/v4/catalog/%s", input.ID)
	catalogs = append(catalogs, input)
	muCat.Unlock()

	go dispatchEvent(buildEvent("CatalogCreateEvent", input))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", input.Href)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(input); err != nil {
		log.Printf("unable to encode response: %v", err)
	}
}

func handleUpdateCatalog(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := r.PathValue("id")
	muCat.Lock()
	index, catalogPtr := findCatalog(id)
	if index == -1 {
		muCat.Unlock()
		writeError(w, http.StatusNotFound, "404", "invalid ID", "ID doesn't exist in the store")
		return
	}
	catalog := *catalogPtr
	muCat.Unlock()

	existing, err := json.Marshal(catalog)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "500", "internal error", "unable to marshal catalog")
		return
	}
	var original map[string]any
	if err := json.Unmarshal(existing, &original); err != nil {
		writeError(w, http.StatusInternalServerError, "500", "internal error", "unable to unmarshal catalog")
		return
	}

	var overlay map[string]any
	if err := json.NewDecoder(r.Body).Decode(&overlay); err != nil {
		writeError(w, http.StatusBadRequest, "400", "malformed request", "body is not valid JSON")
		return
	}

	for key, value := range overlay {
		original[key] = value
	}

	newCat, err := json.Marshal(original)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "500", "internal error", "unable to marshal updated catalog")
		return
	}
	var updated Catalog
	if err := json.Unmarshal(newCat, &updated); err != nil {
		writeError(w, http.StatusInternalServerError, "500", "internal error", "unable to unmarshal updated catalog")
		return
	}

	updated.LastUpdate = time.Now().UTC().Format(time.RFC3339)
	updated.ID = catalog.ID
	updated.Href = catalog.Href

	muCat.Lock()
	index, catalogPtr = findCatalog(id)
	if index == -1 {
		muCat.Unlock()
		writeError(w, http.StatusNotFound, "404", "invalid ID", "ID no longer exists in the store")
		return
	}
	catalogs[index] = updated
	muCat.Unlock()

	go dispatchEvent(buildEvent("CatalogAttributeValueChangeEvent", updated))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updated); err != nil {
		log.Printf("encode error: %v", err)
		return
	}
}

func handleDeleteCatalog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	muCat.Lock()
	index, catalog := findCatalog(id)
	if index == -1 {
		muCat.Unlock()
		writeError(w, http.StatusNotFound, "404", "not found", "ID doesn't exist in the store")
		return
	}
	deleted := *catalog
	catalogs = slices.Delete(catalogs, index, index+1)
	muCat.Unlock()

	go dispatchEvent(buildEvent("CatalogDeleteEvent", deleted))
	w.WriteHeader(http.StatusNoContent)
}

func handleRegisterHub(w http.ResponseWriter, r *http.Request) {
	var body HubSubscription
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "400", "bad request", "body is not valid JSON")
		return
	}
	if body.Callback == "" {
		writeError(w, http.StatusBadRequest, "400", "bad request", "empty callback")
		return
	}

	body.ID = generateID()
	subsMu.Lock()
	if slices.ContainsFunc(subscriptions, func(s HubSubscription) bool { return s.Callback == body.Callback }) {
		subsMu.Unlock()
		writeError(w, http.StatusConflict, "409", "callback conflict", "callback already registered")
		return
	}
	subscriptions = append(subscriptions, body)
	subsMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/tmf-api/hub/%s", body.ID))
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

	if originalLen == finalLen {
		writeError(w, http.StatusNotFound, "404", "not found", "hub subscription ID does not exist")
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
