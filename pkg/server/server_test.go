package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeStatusEndpoint struct {
	status int
}

func (h *fakeStatusEndpoint) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(h.status)
}

type fakeCompositionEndpoint struct {
	status int
}

func (h *fakeCompositionEndpoint) GetComposition(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(h.status)
}

func TestPingEndpoint(t *testing.T) {
	t.Parallel()

	s := New(
		":0",
		&fakeStatusEndpoint{status: http.StatusOK},
		&fakeCompositionEndpoint{status: http.StatusOK},
		Options{
			CompositionAPIKey:             "test-key",
			CompositionRateLimitPerMinute: 60,
		},
	)
	req, err := http.NewRequest(http.MethodGet, "/ping", http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	rr := httptest.NewRecorder()
	s.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusOK)
	}
}

func TestGetCompositionEndpoint(t *testing.T) {
	t.Parallel()

	s := New(
		":0",
		&fakeStatusEndpoint{status: http.StatusOK},
		&fakeCompositionEndpoint{status: http.StatusCreated},
		Options{
			CompositionAPIKey:             "test-key",
			CompositionRateLimitPerMinute: 60,
		},
	)
	req, err := http.NewRequest(
		http.MethodPost,
		"/get-composition",
		strings.NewReader(`{}`),
	)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key")
	rr := httptest.NewRecorder()
	s.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusCreated)
	}
}

func TestGetCompositionEndpoint_Unauthorized(t *testing.T) {
	t.Parallel()

	s := New(
		":0",
		&fakeStatusEndpoint{status: http.StatusOK},
		&fakeCompositionEndpoint{status: http.StatusCreated},
		Options{
			CompositionAPIKey:             "test-key",
			CompositionRateLimitPerMinute: 60,
		},
	)

	req, err := http.NewRequest(
		http.MethodPost,
		"/get-composition",
		strings.NewReader(`{}`),
	)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusUnauthorized)
	}
}

func TestGetCompositionEndpoint_RateLimited(t *testing.T) {
	t.Parallel()

	s := New(
		":0",
		&fakeStatusEndpoint{status: http.StatusOK},
		&fakeCompositionEndpoint{status: http.StatusCreated},
		Options{
			CompositionAPIKey:             "test-key",
			CompositionRateLimitPerMinute: 1,
		},
	)

	makeRequest := func() *httptest.ResponseRecorder {
		req, err := http.NewRequest(
			http.MethodPost,
			"/get-composition",
			strings.NewReader(`{}`),
		)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")
		req.RemoteAddr = "203.0.113.10:1234"
		rr := httptest.NewRecorder()
		s.Handler.ServeHTTP(rr, req)
		return rr
	}

	first := makeRequest()
	if first.Code != http.StatusCreated {
		t.Fatalf("first status=%d want=%d", first.Code, http.StatusCreated)
	}

	second := makeRequest()
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second status=%d want=%d", second.Code, http.StatusTooManyRequests)
	}
}
