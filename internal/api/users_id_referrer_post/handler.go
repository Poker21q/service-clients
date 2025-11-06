package users_id_referrer_post

import (
	"context"
	"encoding/json"
	"net/http"

	"service-boilerplate-go/internal/generated/api"
	"service-boilerplate-go/internal/pkg/middleware/jwtauth"
	"service-boilerplate-go/internal/pkg/response"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Logger interface {
	Error(ctx context.Context, msg string)
	Warn(ctx context.Context, msg string)
	Info(ctx context.Context, msg string)

	WithFields(ctx context.Context, fields map[string]any) context.Context
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
		h.logger.Warn(ctx, "unauthorized: no user id in context")
		response.ErrorStatus(w, http.StatusUnauthorized)
		return
	}

	userIDStr := mux.Vars(r)["id"]
	ctx = h.logger.WithFields(ctx, map[string]any{
		"user_id_path":  userIDStr,
		"user_id_token": userIDStrCtx,
	})

	if userIDStrCtx != userIDStr {
		h.logger.Warn(ctx, "unauthorized: token user id does not match path user id")
		response.ErrorStatus(w, http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx = h.logger.WithFields(ctx, map[string]any{
			"error": err.Error(),
		})
		h.logger.Warn(ctx, "invalid user id format")
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	var req api.ReferrerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx = h.logger.WithFields(ctx, map[string]any{
			"error": err.Error(),
		})
		h.logger.Warn(ctx, "failed to decode json body")
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	if req.ReferrerId == uuid.Nil {
		h.logger.Warn(ctx, "empty referrer id")
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	ctx = h.logger.WithFields(ctx, map[string]any{
		"referrer_id": req.ReferrerId,
	})

	if err := h.service.InputReferrer(ctx, userID, req.ReferrerId); err != nil {
		ctx = h.logger.WithFields(ctx, map[string]any{
			"error": err.Error(),
		})
		h.logger.Error(ctx, "failed to input referrer")
		response.ErrorDomain(w, err)
		return
	}

	h.logger.Info(ctx, "referrer input successful")
	response.OkJSON(w, api.ReferrerResponse{Status: "ok"})
}
