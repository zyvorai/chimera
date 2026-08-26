package gateway

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type RequestEvent struct {
	ID         uint64 `json:"id"`
	At         string `json:"at"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	DurationMS int64  `json:"duration_ms"`
	Bytes      int64  `json:"bytes"`
	Category   string `json:"category"`
	Partial    bool   `json:"partial"`
}

type ActivityItem struct {
	Name    string  `json:"name"`
	Count   uint64  `json:"count"`
	Percent float64 `json:"percent"`
}

type TelemetrySnapshot struct {
	Requests         uint64         `json:"requests"`
	Errors           uint64         `json:"errors"`
	ErrorRate        float64        `json:"error_rate_pct"`
	ActiveRequests   int64          `json:"active_requests"`
	ActiveSessions   int            `json:"active_sessions"`
	BytesTransferred uint64         `json:"bytes_transferred"`
	Exports          uint64         `json:"exports"`
	AverageResponse  float64        `json:"average_response_ms"`
	Recent           []RequestEvent `json:"recent"`
	Activity         []ActivityItem `json:"activity"`
}

type telemetry struct {
	mu sync.RWMutex

	seq      uint64
	requests uint64
	errors   uint64
	bytes    uint64
	exports  uint64
	totalMS  uint64
	active   int64
	recent   []RequestEvent
	activity map[string]uint64
	sessions map[string]time.Time
}

func newTelemetry() *telemetry {
	return &telemetry{
		recent:   make([]RequestEvent, 0, 80),
		activity: make(map[string]uint64),
		sessions: make(map[string]time.Time),
	}
}

func (t *telemetry) begin(r *http.Request) time.Time {
	atomic.AddInt64(&t.active, 1)
	if c, err := r.Cookie("vmware_soap_session"); err == nil && c.Value != "" {
		t.mu.Lock()
		t.sessions[c.Value] = time.Now()
		t.mu.Unlock()
	}
	return time.Now()
}

func (t *telemetry) finish(r *http.Request, status int, bytes int64, started time.Time) {
	atomic.AddInt64(&t.active, -1)
	if status == 0 {
		status = http.StatusOK
	}
	d := time.Since(started)
	ms := d.Milliseconds()
	if ms == 0 && d > 0 {
		ms = 1
	}
	category := requestCategory(r)
	partial := r.Header.Get("Range") != "" || status == http.StatusPartialContent

	t.mu.Lock()
	defer t.mu.Unlock()
	t.seq++
	t.requests++
	if status >= 400 {
		t.errors++
	}
	if bytes > 0 {
		t.bytes += uint64(bytes)
	}
	if ms > 0 {
		t.totalMS += uint64(ms)
	}
	if strings.HasPrefix(r.URL.Path, "/chimera-nfc/") && r.Method == http.MethodGet && !partial {
		t.exports++
	}
	t.activity[category]++
	e := RequestEvent{
		ID: t.seq, At: time.Now().Format(time.RFC3339Nano), Method: r.Method, Path: compactPath(r.URL.Path),
		Status: status, DurationMS: ms, Bytes: bytes, Category: category, Partial: partial,
	}
	t.recent = append([]RequestEvent{e}, t.recent...)
	if len(t.recent) > 80 {
		t.recent = t.recent[:80]
	}

	cutoff := time.Now().Add(-30 * time.Minute)
	for k, seen := range t.sessions {
		if seen.Before(cutoff) {
			delete(t.sessions, k)
		}
	}
}

func (t *telemetry) snapshot() TelemetrySnapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := TelemetrySnapshot{
		Requests: t.requests, Errors: t.errors, BytesTransferred: t.bytes, Exports: t.exports,
		ActiveRequests: atomic.LoadInt64(&t.active), ActiveSessions: len(t.sessions),
		Recent: append([]RequestEvent(nil), t.recent...),
	}
	if t.requests > 0 {
		out.ErrorRate = float64(t.errors) * 100 / float64(t.requests)
		out.AverageResponse = float64(t.totalMS) / float64(t.requests)
	}
	items := make([]ActivityItem, 0, len(t.activity))
	for name, count := range t.activity {
		pct := 0.0
		if t.requests > 0 {
			pct = float64(count) * 100 / float64(t.requests)
		}
		items = append(items, ActivityItem{Name: name, Count: count, Percent: pct})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Name < items[j].Name
		}
		return items[i].Count > items[j].Count
	})
	if len(items) > 6 {
		items = items[:6]
	}
	out.Activity = items
	return out
}

func requestCategory(r *http.Request) string {
	p := r.URL.Path
	switch {
	case strings.HasPrefix(p, "/chimera-nfc/") || strings.HasPrefix(p, "/nfc/"):
		if r.Header.Get("Range") != "" {
			return "NFC Resume"
		}
		return "NFC Download"
	case p == "/sdk" && r.Method == http.MethodPost:
		return "vSphere SOAP"
	case p == "/sdk":
		return "vSphere SDK"
	default:
		return "Other"
	}
}

func compactPath(p string) string {
	if len(p) <= 72 {
		return p
	}
	return p[:34] + "…" + p[len(p)-34:]
}

type captureWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (w *captureWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *captureWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += int64(n)
	return n, err
}

func (w *captureWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *captureWriter) Flush() {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
