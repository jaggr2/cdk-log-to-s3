package main

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TelemetryBatch is one element of the JSON array the Telemetry API POSTs to
// the local HTTP listener of this extension.
type TelemetryBatch struct {
	Time   string          `json:"time"`
	Type   string          `json:"type"`
	Record json.RawMessage `json:"record"`
}

// StructuredFields are the sentinel keys emitted by the companion logger
// package (@jaggr2/log-to-s3-logger). __log_level is the discriminator: a JSON
// line without it is treated as an ordinary text log, not as a structured one.
type StructuredFields struct {
	LogLevel      string          `json:"__log_level"`
	Source        string          `json:"__source"`
	CorrelationID string          `json:"__correlation_id"`
	Caller        string          `json:"__caller"`
	Message       string          `json:"message"`
	Context       json.RawMessage `json:"context,omitempty"`
	StackTrace    string          `json:"stackTrace,omitempty"`
}

// cwPrefixRe matches the "TIMESTAMP<tab>REQUEST_ID<tab>LEVEL<tab>" prefix the
// runtime prepends in the default (text) log format.
//
// This is anchored on purpose. The previous implementation searched for the
// last tab-brace pair in the record, which both mis-fired on plain text
// containing a tab followed by a brace and threw away the request id.
var cwPrefixRe = regexp.MustCompile(
	`^(\d{4}-\d{2}-\d{2}T[0-9:.]+Z)\t([0-9a-fA-F-]{36})\t(?:([A-Z]+)\t)?`)

type Parser struct {
	cfg *Config
	now func() time.Time
}

func NewParser(cfg *Config, now func() time.Time) *Parser {
	if now == nil {
		now = time.Now
	}
	return &Parser{cfg: cfg, now: now}
}

// ParseBatch converts one telemetry event into zero or more log entries.
// The second return value asks the caller to flush: the runtime has finished
// the invocation and the sandbox may be frozen at any moment.
func (p *Parser) ParseBatch(b TelemetryBatch) ([]LogEntry, bool) {
	switch {
	case b.Type == "function":
		return p.parseFunction(b), false

	// The Telemetry API reports fully qualified platform types such as
	// "platform.report" and "platform.runtimeDone". The older Logs API used a
	// bare "platform" with the specific type nested in the record; both shapes
	// are accepted so the extension is not tied to one API generation.
	case b.Type == "platform.runtimeDone":
		return nil, true
	case b.Type == "platform.report":
		return p.parsePlatformReport(b, b.Record), false
	case strings.HasPrefix(b.Type, "platform."):
		return nil, false
	case b.Type == "platform":
		var nested struct {
			Type   string          `json:"type"`
			Record json.RawMessage `json:"record"`
		}
		if err := json.Unmarshal(b.Record, &nested); err != nil {
			return nil, false
		}
		switch nested.Type {
		case "runtimeDone":
			return nil, true
		case "report":
			return p.parsePlatformReport(b, nested.Record), false
		}
		return nil, false

	default:
		// "extension" records are deliberately never subscribed to; if one
		// ever arrived it would be the stdout of this extension feeding back
		// into its own buffer.
		return nil, false
	}
}

func (p *Parser) parseFunction(b TelemetryBatch) []LogEntry {
	entry := LogEntry{
		Timestamp:    p.timestampOf(b),
		Level:        "INFO",
		FunctionName: p.cfg.FunctionName,
	}

	// Text log format: record is a JSON string.
	var text string
	if err := json.Unmarshal(b.Record, &text); err == nil {
		p.fillFromText(&entry, text)
		return []LogEntry{entry}
	}

	// JSON log format (LAMBDA_LOG_FORMAT=JSON): record is an object. The old
	// implementation fell back to the raw bytes here and produced garbage.
	var obj struct {
		Timestamp string          `json:"timestamp"`
		Level     string          `json:"level"`
		RequestID string          `json:"requestId"`
		Message   json.RawMessage `json:"message"`
	}
	if err := json.Unmarshal(b.Record, &obj); err == nil {
		if obj.Timestamp != "" {
			entry.Timestamp = obj.Timestamp
		}
		if obj.Level != "" {
			entry.Level = strings.ToUpper(obj.Level)
		}
		entry.RequestID = obj.RequestID

		// message may be a plain string, a string containing JSON, or an
		// already-decoded object.
		var msg string
		if err := json.Unmarshal(obj.Message, &msg); err == nil {
			if !applyStructuredFields(&entry, msg) {
				entry.Message = strings.TrimSpace(msg)
			}
		} else if len(obj.Message) > 0 {
			if !applyStructuredFields(&entry, string(obj.Message)) {
				entry.Message = strings.TrimSpace(string(obj.Message))
			}
		}
		return []LogEntry{entry}
	}

	entry.Message = strings.TrimSpace(string(b.Record))
	return []LogEntry{entry}
}

func (p *Parser) fillFromText(entry *LogEntry, record string) {
	body := record
	if m := cwPrefixRe.FindStringSubmatch(record); m != nil {
		body = record[len(m[0]):]
		entry.Timestamp = m[1]
		entry.RequestID = m[2]
		if m[3] != "" {
			entry.Level = m[3]
		}
	}

	body = strings.TrimSpace(body)
	if applyStructuredFields(entry, body) {
		return
	}

	// Fallback for a prefix shape the regex does not recognise. It is only
	// accepted when the tail really is a structured record: the previous
	// implementation cut at the last tab-brace pair unconditionally and so
	// truncated any plain-text line that happened to contain one.
	if idx := strings.Index(record, "\t{"); idx >= 0 {
		if tail := strings.TrimSpace(record[idx+1:]); applyStructuredFields(entry, tail) {
			return
		}
	}

	if entry.Level == "INFO" && containsErrorWord(record) {
		entry.Level = "ERROR"
	}
	entry.Message = body
}

// applyStructuredFields decodes the JSON contract of the companion logger. It
// returns false for anything that is not a structured record, so the caller
// can fall back to plain-text handling.
func applyStructuredFields(entry *LogEntry, body string) bool {
	body = strings.TrimSpace(body)
	if !strings.HasPrefix(body, "{") {
		return false
	}
	var sf StructuredFields
	if err := json.Unmarshal([]byte(body), &sf); err != nil || sf.LogLevel == "" {
		return false
	}
	entry.Level = strings.ToUpper(sf.LogLevel)
	entry.Source = sf.Source
	entry.CorrelationID = sf.CorrelationID
	entry.Message = sf.Message
	entry.Caller = strPtr(sf.Caller)
	if len(sf.Context) > 0 && string(sf.Context) != "null" {
		entry.Context = strPtr(string(sf.Context))
	}
	entry.StackTrace = strPtr(sf.StackTrace)
	return true
}

func containsErrorWord(s string) bool {
	return strings.Contains(s, "ERROR") ||
		strings.Contains(s, "Error") ||
		strings.Contains(s, "error")
}

func (p *Parser) parsePlatformReport(b TelemetryBatch, record json.RawMessage) []LogEntry {
	if !p.cfg.IncludePlatformReport {
		return nil
	}

	var report struct {
		RequestID string `json:"requestId"`
		Metrics   struct {
			DurationMs       float64 `json:"durationMs"`
			BilledDurationMs float64 `json:"billedDurationMs"`
			MemorySizeMB     float64 `json:"memorySizeMB"`
			MaxMemoryUsedMB  float64 `json:"maxMemoryUsedMB"`
			InitDurationMs   float64 `json:"initDurationMs"`
		} `json:"metrics"`
	}
	if err := json.Unmarshal(record, &report); err != nil {
		return nil
	}

	m := report.Metrics
	ctx, err := json.Marshal(map[string]float64{
		"durationMs":     m.DurationMs,
		"billedMs":       m.BilledDurationMs,
		"memoryUsedMB":   m.MaxMemoryUsedMB,
		"memorySizeMB":   m.MemorySizeMB,
		"initDurationMs": m.InitDurationMs,
	})
	if err != nil {
		return nil
	}

	entry := LogEntry{
		Timestamp:    p.timestampOf(b),
		Level:        "INFO",
		Source:       "platform",
		RequestID:    report.RequestID,
		FunctionName: p.cfg.FunctionName,
		Message: formatReport(m.DurationMs, m.BilledDurationMs,
			m.MaxMemoryUsedMB, m.MemorySizeMB),
		Context: strPtr(string(ctx)),
	}
	return []LogEntry{entry}
}

// timestampOf prefers the timestamp the Telemetry API reports for the event
// over the clock of this extension, which can be seconds later.
func (p *Parser) timestampOf(b TelemetryBatch) string {
	if b.Time != "" {
		return b.Time
	}
	return p.now().UTC().Format(time.RFC3339Nano)
}

func formatReport(durationMs, billedMs, usedMB, sizeMB float64) string {
	return "Invocation: " + trimFloat(durationMs) + "ms (billed " +
		trimFloat(billedMs) + "ms), memory " + trimFloat(usedMB) + "/" +
		trimFloat(sizeMB) + " MB"
}

func trimFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
