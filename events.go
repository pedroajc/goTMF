// events.go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	subscriptions []HubSubscription
	subsMu        sync.RWMutex
)

func buildEvent(eventType string, catalog Catalog) Event {
	var event Event
	event.ID = generateID()
	event.EventID = event.ID
	event.EventTime = time.Now().UTC().Format(time.RFC3339)
	event.EventType = eventType
	event.Event.Catalog = &catalog

	return event
}

func dispatchEvent(event Event) {
	subsMu.Lock()
	snapshot := make([]HubSubscription, len(subscriptions))
	copy(snapshot, subscriptions)
	subsMu.Unlock()

	for _, subs := range snapshot {
		if subs.Query != "" && !strings.Contains(subs.Query, event.EventType) {
			continue
		}
		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		resp, err := makeRequest(ctx, subs.Callback, http.MethodPost, bytes.NewReader(data))
		cancel()
		if err != nil {
			log.Printf("error dispatching event: %v", err)
			continue
		}
		resp.Body.Close()
		log.Printf("event %s dispatched", event.EventID)
	}
}
