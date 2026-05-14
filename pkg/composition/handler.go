package composition

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

const maxUploadBytesDefault = 10 << 20

type CompositionService interface {
	GetComposition(ctx context.Context, req Request) (Result, error)
}

type Handler struct {
	service        CompositionService
	maxUploadBytes int64
}

type getCompositionJSONRequest struct {
	ProductName string `json:"product_name"`
}

func NewHandler(service CompositionService) *Handler {
	return NewHandlerWithMaxUploadBytes(service, maxUploadBytesDefault)
}

func NewHandlerWithMaxUploadBytes(service CompositionService, maxUploadBytes int64) *Handler {
	if maxUploadBytes <= 0 {
		maxUploadBytes = maxUploadBytesDefault
	}

	return &Handler{
		service:        service,
		maxUploadBytes: maxUploadBytes,
	}
}

func (h *Handler) GetComposition(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes)

	request, err := h.readRequest(r)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			h.writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.GetComposition(r.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmptyInput), errors.Is(err, ErrConflictingData):
			h.writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrInvalidImageContentType):
			h.writeError(w, http.StatusUnsupportedMediaType, err.Error())
		case errors.Is(err, ErrNoComposition):
			h.writeError(w, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, ErrChatGPTNotConfigured):
			h.writeError(w, http.StatusServiceUnavailable, err.Error())
		case errors.Is(err, ErrChatGPTInvalidResponse):
			h.writeError(w, http.StatusBadGateway, err.Error())
		default:
			h.writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

func (h *Handler) readRequest(r *http.Request) (Request, error) {
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		return h.readJSONRequest(r)
	}

	return h.readMultipartRequest(r)
}

func (h *Handler) readJSONRequest(r *http.Request) (Request, error) {
	var payload getCompositionJSONRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return Request{}, errors.New("invalid json body")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Request{}, errors.New("invalid json body")
	}

	return Request{
		ProductName: payload.ProductName,
	}, nil
}

func (h *Handler) readMultipartRequest(r *http.Request) (Request, error) {
	if err := r.ParseMultipartForm(h.maxUploadBytes); err != nil {
		return Request{}, errors.New("invalid multipart form body")
	}

	request := Request{
		ProductName: r.FormValue("product_name"),
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return request, nil
		}
		return Request{}, errors.New("invalid image file")
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		return Request{}, errors.New("failed to read image file")
	}

	request.ImageBytes = imageBytes
	request.ImageName = header.Filename
	return request, nil
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
