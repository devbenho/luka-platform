package errors

import (
	"encoding/json"
	"log"
	"net/http"
)

func ErrorHandler(h func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		if err := h(w, r); err != nil {
			apiErr := MapErrorToHTTP(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(apiErr.Status)
			json.NewEncoder(w).Encode(apiErr)
		}
	}
}
