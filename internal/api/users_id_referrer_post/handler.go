package users_id_referrer_post

import (
	"context"
	"encoding/json"
	"net/http"

	"service-boilerplate-go/internal/generated/api"
	"service-boilerplate-go/internal/pkg/middlewares/jwtauth"
	"service-boilerplate-go/internal/pkg/response"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Logger interface {
	Error(ctx context.Context, msg string)
}

type Service interface {
	InputReferrer(ctx context.Context, userID, referrerID uuid.UUID) error
}

type Handler struct {
	logger  Logger
	service Service
}

func New(logger Logger, service Service) *Handler {
	return &Handler{logger: logger, service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDStrCtx, ok := jwtauth.UserIDFromContext(ctx)
	if !ok {
		response.ErrorStatus(w, http.StatusUnauthorized)
		return
	}

	userIDStr := mux.Vars(r)["id"]
	if userIDStrCtx != userIDStr {
		response.ErrorStatus(w, http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	var req api.ReferrerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	if req.ReferrerId == uuid.Nil {
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	if err := h.service.InputReferrer(ctx, userID, req.ReferrerId); err != nil {
		response.ErrorDomain(w, err)
		return
	}

	response.OkJSON(w, api.ReferrerResponse{Status: "ok"})
}
