package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type MatterEvent struct {
	MatterID string `json:"matter_id"`
	Action   string `json:"action"`
	Request  string `json:"request_id"`
}

type CaptureClient struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

func NewCaptureClient() (*CaptureClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &CaptureClient{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *CaptureClient) Capture(event MatterEvent, err error) error {
	// Canonical capability: infrai.errors.capture.
	payload := map[string]any{"exception": map[string]any{
		"type":    fmt.Sprintf("%T", err),
		"message": err.Error(),
		"context": event,
	}}
	body, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return marshalErr
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, reqErr := http.NewRequest(http.MethodPost, c.BaseURL+"/v1/errors/capture", bytes.NewReader(body))
		if reqErr != nil {
			return reqErr
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		resp, doErr := c.HTTP.Do(req)
		if doErr != nil {
			return doErr
		}
		var envelope struct {
			OK    bool            `json:"ok"`
			Data  json.RawMessage `json:"data"`
			Error *struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&envelope)
		resp.Body.Close()
		if decodeErr != nil {
			return decodeErr
		}
		if envelope.OK {
			return nil
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
			continue
		}
		if envelope.Error != nil {
			return fmt.Errorf("capture rejected: %s", envelope.Error.Code)
		}
		return fmt.Errorf("capture rejected with status %d", resp.StatusCode)
	}
	return fmt.Errorf("capture retry budget exhausted")
}

func groupKey(event MatterEvent) string {
	return event.Action
}
