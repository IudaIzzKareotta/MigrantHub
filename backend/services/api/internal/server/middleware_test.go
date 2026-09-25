package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func decodeRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("decode log record %q: %v", buf.String(), err)
	}
	return record
}

func TestResponseWriterWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec}

	rw.WriteHeader(http.StatusCreated)

	if rw.statusCode != http.StatusCreated {
		t.Errorf("statusCode = %d, want %d", rw.statusCode, http.StatusCreated)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("underlying code = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestResponseWriterWriteHeaderKeepsFirstStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec}

	rw.WriteHeader(http.StatusTeapot)
	rw.WriteHeader(http.StatusInternalServerError)

	if rw.statusCode != http.StatusTeapot {
		t.Errorf("statusCode = %d, want %d", rw.statusCode, http.StatusTeapot)
	}
	if rec.Code != http.StatusTeapot {
		t.Errorf("underlying code = %d, want %d", rec.Code, http.StatusTeapot)
	}
}

func TestResponseWriterWriteDefaultsToOK(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec}

	n, err := rw.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != len("hello") {
		t.Errorf("Write() n = %d, want %d", n, len("hello"))
	}
	if rw.statusCode != http.StatusOK {
		t.Errorf("statusCode = %d, want %d", rw.statusCode, http.StatusOK)
	}
	if got := rec.Body.String(); got != "hello" {
		t.Errorf("body = %q, want %q", got, "hello")
	}
}

func TestResponseWriterWriteAfterExplicitStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec}

	rw.WriteHeader(http.StatusAccepted)
	if _, err := rw.Write([]byte("body")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if rw.statusCode != http.StatusAccepted {
		t.Errorf("statusCode = %d, want %d", rw.statusCode, http.StatusAccepted)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("underlying code = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if got := rec.Body.String(); got != "body" {
		t.Errorf("body = %q, want %q", got, "body")
	}
}

func TestRequestLoggerCallsNext(t *testing.T) {
	var buf bytes.Buffer

	called := false
	handler := RequestLogger(newTestLogger(&buf))(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusNoContent)
		}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if !called {
		t.Fatal("next handler was not called")
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("response code = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestRequestLoggerPassesThroughRequestAndResponse(t *testing.T) {
	var buf bytes.Buffer

	var gotMethod, gotPath string
	var gotBody []byte

	handler := RequestLogger(newTestLogger(&buf))(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			gotBody, _ = io.ReadAll(r.Body)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":1}`))
		}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"name":"ann"}`))
	handler.ServeHTTP(rec, req)

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/users" {
		t.Errorf("path = %q, want %q", gotPath, "/users")
	}
	if string(gotBody) != `{"name":"ann"}` {
		t.Errorf("request body = %q, want %q", gotBody, `{"name":"ann"}`)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("response code = %d, want %d", rec.Code, http.StatusCreated)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}
	if got := rec.Body.String(); got != `{"id":1}` {
		t.Errorf("response body = %q, want %q", got, `{"id":1}`)
	}
}

func TestRequestLoggerLogsRequestFields(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		target     string
		handler    http.HandlerFunc
		wantStatus float64
	}{
		{
			name:   "explicit status",
			method: http.MethodDelete,
			target: "/users/1",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "implicit status from Write",
			method: http.MethodGet,
			target: "/health",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"status":"ok"}`))
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "error status",
			method: http.MethodGet,
			target: "/boom",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "boom", http.StatusInternalServerError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			handler := RequestLogger(newTestLogger(&buf))(tt.handler)
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(tt.method, tt.target, nil))

			record := decodeRecord(t, &buf)

			if got := record["msg"]; got != "HTTP request" {
				t.Errorf("msg = %v, want %q", got, "HTTP request")
			}
			if got := record["method"]; got != tt.method {
				t.Errorf("method = %v, want %q", got, tt.method)
			}
			if got := record["path"]; got != tt.target {
				t.Errorf("path = %v, want %q", got, tt.target)
			}
			if got := record["status"]; got != tt.wantStatus {
				t.Errorf("status = %v, want %v", got, tt.wantStatus)
			}

			duration, ok := record["duration"].(float64)
			if !ok {
				t.Fatalf("duration = %v (%T), want a number", record["duration"], record["duration"])
			}
			if duration < 0 {
				t.Errorf("duration = %v, want >= 0", duration)
			}
		})
	}
}

func TestRequestLoggerLogsPathWithoutQuery(t *testing.T) {
	var buf bytes.Buffer

	handler := RequestLogger(newTestLogger(&buf))(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/search?q=visa&page=2", nil))

	record := decodeRecord(t, &buf)
	if got := record["path"]; got != "/search" {
		t.Errorf("path = %v, want %q", got, "/search")
	}
}

func TestRequestLoggerLogsZeroStatusWhenHandlerNeverWrites(t *testing.T) {
	var buf bytes.Buffer

	handler := RequestLogger(newTestLogger(&buf))(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	record := decodeRecord(t, &buf)
	if got := record["status"]; got != float64(0) {
		t.Errorf("status = %v, want 0", got)
	}
}
