package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// The destination host is the single most breakage-prone value in the
// subscription: Lambda accepts a bad one and then fails every delivery
// silently, so nothing surfaces until an end-to-end run finds no data.
func TestSubscribeDestinationUsesSandboxLocaldomain(t *testing.T) {
	req := buildSubscribeRequest(&Config{TelemetryPort: "2020"})

	if req.Destination.URI != "http://sandbox.localdomain:2020" {
		t.Errorf("URI = %q, want http://sandbox.localdomain:2020", req.Destination.URI)
	}
	if req.Destination.Protocol != "HTTP" {
		t.Errorf("Protocol = %q, want HTTP", req.Destination.Protocol)
	}
}

func TestSubscribeHonoursThePort(t *testing.T) {
	req := buildSubscribeRequest(&Config{TelemetryPort: "3999"})
	if !strings.HasSuffix(req.Destination.URI, ":3999") {
		t.Errorf("URI = %q, want it to end in :3999", req.Destination.URI)
	}
}

// Subscribing to "extension" would feed this extension's own stdout back into
// its own buffer.
func TestSubscribeNeverAsksForExtensionLogs(t *testing.T) {
	req := buildSubscribeRequest(&Config{TelemetryPort: "2020"})

	for _, typ := range req.Types {
		if typ == "extension" {
			t.Fatal("must not subscribe to the extension stream")
		}
	}
	if len(req.Types) != 2 {
		t.Errorf("Types = %v, want platform + function", req.Types)
	}
}

// Buffering values must stay inside the documented limits, or Lambda rejects
// the subscription.
func TestSubscribeBufferingWithinDocumentedLimits(t *testing.T) {
	b := buildSubscribeRequest(&Config{TelemetryPort: "2020"}).Buffering

	if b.MaxItems < 1000 || b.MaxItems > 10000 {
		t.Errorf("MaxItems = %d, outside 1000..10000", b.MaxItems)
	}
	if b.MaxBytes < 262144 || b.MaxBytes > 1048576 {
		t.Errorf("MaxBytes = %d, outside 262144..1048576", b.MaxBytes)
	}
	if b.TimeoutMs < 25 || b.TimeoutMs > 30000 {
		t.Errorf("TimeoutMs = %d, outside 25..30000", b.TimeoutMs)
	}
}

func TestSubscribeRequestMarshals(t *testing.T) {
	data, err := json.Marshal(buildSubscribeRequest(&Config{TelemetryPort: "2020"}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"URI":"http://sandbox.localdomain:2020"`) {
		t.Errorf("marshalled body = %s", data)
	}
}
