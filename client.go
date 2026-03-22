// client.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

func makeRequest(ctx context.Context, url string, method string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error setting up the request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making the request: %w", err)
	}
	return resp, nil
}

func fetchRemoteCatalogs(ctx context.Context, nodeURL string) ([]Catalog, error) {
	resp, err := makeRequest(ctx, nodeURL+"/catalogManagement/v4/catalog", http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code %d instead of %d", resp.StatusCode, http.StatusOK)
	}

	var catalogs []Catalog
	if err := json.NewDecoder(resp.Body).Decode(&catalogs); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return catalogs, nil
}

func fetchRemoteCatalog(ctx context.Context, nodeURL, id string) (*Catalog, error) {
	resp, err := makeRequest(ctx, nodeURL+"/catalogManagement/v4/catalog/"+id, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code %d instead of %d", resp.StatusCode, http.StatusOK)
	}

	var catalog Catalog
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return &catalog, nil
}
