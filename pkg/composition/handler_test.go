package composition

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeCompositionService struct {
	result Result
	err    error
}

func (s *fakeCompositionService) GetComposition(context.Context, Request) (Result, error) {
	return s.result, s.err
}

func TestGetCompositionJSONRequest(t *testing.T) {
	t.Parallel()

	handler := NewHandler(&fakeCompositionService{
		result: Result{Items: []string{"Water", "Sugar"}},
	})

	req := httptest.NewRequest(http.MethodPost, "/get-composition", strings.NewReader(`{"product_name":"Fanta"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.GetComposition(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusOK)
	}

	var got Result
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Items) != 2 {
		t.Fatalf("items=%v want len=2", got.Items)
	}
}

func TestGetCompositionMultipartRequest(t *testing.T) {
	t.Parallel()

	handler := NewHandler(&fakeCompositionService{
		result: Result{Items: []string{"Salt"}},
	})

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileWriter, err := writer.CreateFormFile("image", "label.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := fileWriter.Write([]byte("image-data")); err != nil {
		t.Fatalf("Write image: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/get-composition", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	handler.GetComposition(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusOK)
	}
}

func TestGetCompositionChatGPTNotConfigured(t *testing.T) {
	t.Parallel()

	handler := NewHandler(&fakeCompositionService{
		err: ErrChatGPTNotConfigured,
	})

	req := httptest.NewRequest(http.MethodPost, "/get-composition", strings.NewReader(`{"product_name":"Fanta"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.GetComposition(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusServiceUnavailable)
	}
}
