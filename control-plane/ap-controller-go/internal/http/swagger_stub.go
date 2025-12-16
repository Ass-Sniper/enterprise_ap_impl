//go:build !swagger
// +build !swagger

package httpapi

import "github.com/go-chi/chi/v5"

func registerSwagger(r chi.Router) {
	// swagger disabled
}
