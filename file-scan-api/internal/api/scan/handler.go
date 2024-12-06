package scan

import (
	"encoding/json"
	"file-scan-api/internal/api/scan/dto"
	"file-scan-api/internal/clamav"
	"fmt"
	"io"
	"net/http"
)

const (
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"
)

type Handler interface {
	ScanFile() http.HandlerFunc
}

type fileHandler struct {
	service clamav.Service
}

func NewFileScanHandler(service clamav.Service) Handler {
	return &fileHandler{service: service}
}

func (h fileHandler) ScanFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("file")
		if err != nil {
			send(w, dto.APIErr{Err: err}, http.StatusBadRequest)

			return
		}

		var reader io.Reader
		reader = file

		resp, err := h.service.ScanFile(reader)
		if err != nil {
			send(w, dto.APIErr{Err: err}, http.StatusInternalServerError)

			return
		}

		send(w, dto.Response{HasVirus: resp.HasVirus, VirusText: resp.VirusText}, http.StatusOK)
	}
}

func send(w http.ResponseWriter, dto any, statusCode int) {
	w.Header().Set(headerContentType, contentTypeJSON)

	body, err := json.Marshal(dto)
	if err != nil {
		statusCode = http.StatusInternalServerError
		body = []byte(fmt.Sprintf(`{"error":"%s"}`, err))
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}
