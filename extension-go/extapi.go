package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// runtimeAPI is the host:port of the Lambda Runtime API, injected by the
// execution environment.
var runtimeAPI = func() string {
	if addr := os.Getenv("AWS_LAMBDA_RUNTIME_API"); addr != "" {
		return addr
	}
	return "127.0.0.1"
}()

var registerBackoff = []time.Duration{
	100 * time.Millisecond,
	300 * time.Millisecond,
	900 * time.Millisecond,
}

type subscribeRequest struct {
	SchemaVersion string          `json:"schemaVersion"`
	Types         []string        `json:"types"`
	Buffering     bufferingConfig `json:"buffering"`
	Destination   destination     `json:"destination"`
}

type bufferingConfig struct {
	MaxItems  int `json:"maxItems"`
	MaxBytes  int `json:"maxBytes"`
	TimeoutMs int `json:"timeoutMs"`
}

type destination struct {
	Protocol string `json:"protocol"`
	URI      string `json:"URI"`
}

// register announces this extension to the Extensions API and returns the
// identifier every later call must carry.
//
// Registration is retried and then fatal, which is the correct behaviour even
// though it looks harsh: Lambda blocks INIT until every binary in
// /opt/extensions has registered, so an extension that "fails gracefully" and
// keeps running would hang the function instead of the other way round.
func register(extensionName string) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= len(registerBackoff); attempt++ {
		if attempt > 0 {
			time.Sleep(registerBackoff[attempt-1])
		}
		id, err := registerOnce(extensionName)
		if err == nil {
			return id, nil
		}
		lastErr = err
	}
	return "", lastErr
}

func registerOnce(extensionName string) (string, error) {
	url := fmt.Sprintf("http://%s/2020-01-01/extension/register", runtimeAPI)
	body, err := json.Marshal(map[string]interface{}{
		"events": []string{"INVOKE", "SHUTDOWN"},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create register request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Lambda-Extension-Name", extensionName)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("register extension: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("register failed: %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}

	id := resp.Header.Get("Lambda-Extension-Identifier")
	if id == "" {
		return "", fmt.Errorf("register response is missing the Lambda-Extension-Identifier header")
	}
	return id, nil
}

// subscribeTelemetry asks the Telemetry API to POST batches to the local
// listener.
//
// Unlike registration this is genuinely non-fatal. If it fails the function
// keeps working and only the S3 copy of the logs is lost, so the caller logs
// the error and continues into the event loop.
func subscribeTelemetry(extensionID string, cfg *Config) error {
	url := fmt.Sprintf("http://%s/2022-07-01/telemetry", runtimeAPI)

	// Deliberately not subscribing to "extension": that stream carries the
	// stdout of this extension, which would feed its own output back into its
	// own buffer.
	sub := subscribeRequest{
		SchemaVersion: "2022-12-13",
		Types:         []string{"platform", "function"},
		Buffering: bufferingConfig{
			MaxItems:  10000,
			MaxBytes:  1024 * 1024,
			TimeoutMs: 1000,
		},
		Destination: destination{
			Protocol: "HTTP",
			URI:      fmt.Sprintf("http://sandbox:%s", cfg.TelemetryPort),
		},
	}

	data, err := json.Marshal(sub)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create subscribe request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Lambda-Extension-Identifier", extensionID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("subscribe telemetry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telemetry subscribe failed: %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}
	return nil
}

// nextEvent blocks until the runtime reports the next lifecycle event.
func nextEvent(extensionID string) (string, error) {
	url := fmt.Sprintf("http://%s/2020-01-01/extension/event/next", runtimeAPI)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create next request: %w", err)
	}
	req.Header.Set("Lambda-Extension-Identifier", extensionID)

	// No client timeout: this call is a long poll that legitimately blocks for
	// the whole idle period between invocations.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("extension next: %w", err)
	}
	defer resp.Body.Close()

	var event struct {
		EventType string `json:"eventType"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
		return "", fmt.Errorf("decode event: %w", err)
	}
	return event.EventType, nil
}
