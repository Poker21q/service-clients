package users_auth_post

import (
	"context"
	"encoding/json"
	"net/http"

	"service-boilerplate-go/internal/generated/api"
	"service-boilerplate-go/internal/pkg/response"

	"github.com/google/uuid"
)

type Logger interface {
	Error(ctx context.Context, msg string)
	Warn(ctx context.Context, msg string)
	Info(ctx context.Context, msg string)

	WithFields(ctx context.Context, fields map[string]any) context.Context
}

type Service interface {
	Auth(ctx context.Context, username, password string) (token string, userID uuid.UUID, err error)
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

	var req api.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx = h.logger.WithFields(ctx, map[string]any{
			"error": err.Error(),
		})
		h.logger.Warn(ctx, "failed to decode json body")
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		h.logger.Warn(ctx, "empty username")
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	ctx = h.logger.WithFields(ctx, map[string]any{
		"username": req.Username,
	})

	if req.Password == "" {
		h.logger.Warn(ctx, "empty password")
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	token, userID, err := h.service.Auth(ctx, req.Username, req.Password)
	if err != nil {
		ctx = h.logger.WithFields(ctx, map[string]any{
			"error": err.Error(),
		})
		h.logger.Error(ctx, "authentication failed")
		response.ErrorDomain(w, err)
		return
	}

	h.logger.Info(ctx, "authentication successful")

	response.OkJSON(w, api.AuthResponse{
		Token:  token,
		UserId: userID,
	})
}
