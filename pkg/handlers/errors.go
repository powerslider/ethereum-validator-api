package handlers

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Code    int    `json:"code"`
}

func writeAPIError(w http.ResponseWriter, status int, message string, err error) {
	apiErr := APIError{
		Code:    status,
		Message: message,
	}

	if err != nil {
		apiErr.Details = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err = json.NewEncoder(w).Encode(apiErr); err != nil {
		return
	}
}
