package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type HTTPBroadcastAdapter struct {
	BroadcastURL string
	// TODO: Authorization: Bearer <Secret>
	// Secret       string
}

func NewHTTPBroadcastAdapter(broadcastURL string) *HTTPBroadcastAdapter {
	return &HTTPBroadcastAdapter{BroadcastURL: broadcastURL}
}

func (a *HTTPBroadcastAdapter) BroadcastRaw(payload interface{}) error {
	log.Printf("Broadcasting %v t %v", payload, a.BroadcastURL)
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling raw broadcast payload: %w", err)
	}
	resp, err := http.Post(a.BroadcastURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error POSTing broadcast: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("error closing broadcast POST response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code POSTing broadcast: %v (%q)", resp.StatusCode, resp.Status)
	}
	return nil
}
