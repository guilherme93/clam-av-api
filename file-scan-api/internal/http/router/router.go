package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"file-scan-api/internal/api/scan"
	"file-scan-api/internal/appcontroller"
)

func NewRouter(container appcontroller.ServiceContainer) http.Handler {
	router := chi.NewRouter()

	scanHandler := scan.NewFileScanHandler(container.ClamService)

	router.Route("/api/v1/", func(r chi.Router) {
		r.Post("/file-scan/", scanHandler.ScanFile())
	},
	)

	return router
}
