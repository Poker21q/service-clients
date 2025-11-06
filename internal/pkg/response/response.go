package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"service-boilerplate-go/internal/service/entities"

	"service-boilerplate-go/internal/generated/api"
)

func OkJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}

func ErrorStatus(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")

	resp := api.ErrorResponse{}

	switch status {
	case http.StatusBadRequest:
		resp.Errors = "bad request"
	case http.StatusUnauthorized:
		resp.Errors = "unauthorized"
	case http.StatusNotFound:
		resp.Errors = "not found"
	default:
		resp.Errors = "internal server error"
		status = http.StatusInternalServerError
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func ErrorDomain(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, entities.ErrUserNotFound), errors.Is(err, entities.ErrTaskNotFound):
		ErrorStatus(w, http.StatusNotFound)
	case errors.Is(err, entities.ErrInvalidCredentials):
		ErrorStatus(w, http.StatusUnauthorized)
	case errors.Is(err, entities.ErrReferrerAlreadySet), errors.Is(err, entities.ErrTaskAlreadyCompleted):
		ErrorStatus(w, http.StatusBadRequest)

	default:
		ErrorStatus(w, http.StatusInternalServerError)
	}
}
