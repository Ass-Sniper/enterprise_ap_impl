//go:build swagger
// +build swagger

package httpapi

import (
	"os"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "ap-controller-go/docs/openapi"
)

func registerSwagger(r chi.Router) {
	// OpenAPI JSON
	r.Get("/openapi.json", httpSwagger.WrapHandler)

	// Swagger UI (optional)
	if os.Getenv("ENABLE_SWAGGER_UI") == "true" {
		r.Get("/swagger/*", httpSwagger.WrapHandler)
	}
}
