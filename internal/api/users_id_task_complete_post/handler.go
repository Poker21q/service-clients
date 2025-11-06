package users_id_task_complete_post

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

// Service теперь принимает metadata
type Service interface {
	CompleteTask(ctx context.Context, userID, taskID uuid.UUID, metadata map[string]string) error
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

	// получаем userID из JWT
	userIDStrCtx, ok := jwtauth.UserIDFromContext(ctx)
	if !ok {
		h.logger.Warn(ctx, "unauthorized: no user id in context")
		response.ErrorStatus(w, http.StatusUnauthorized)
		return
	}

	// получаем userID из пути
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

	var req api.TaskCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx = h.logger.WithFields(ctx, map[string]any{
			"error": err.Error(),
		})
		h.logger.Warn(ctx, "failed to decode json body")
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	if req.TaskId == uuid.Nil {
		h.logger.Warn(ctx, "empty task id")
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	ctx = h.logger.WithFields(ctx, map[string]any{
		"task_id":  req.TaskId,
		"metadata": req.Metadata,
	})

	// используем метадату, если есть
	var metadata map[string]string
	if req.Metadata != nil {
		metadata = *req.Metadata
	}

	if err := h.service.CompleteTask(ctx, userID, req.TaskId, metadata); err != nil {
		ctx = h.logger.WithFields(ctx, map[string]any{
			"error": err.Error(),
		})
		h.logger.Error(ctx, "failed to complete task")
		response.ErrorDomain(w, err)
		return
	}

	h.logger.Info(ctx, "task completed successfully")
	response.OkJSON(w, api.TaskCompleteResponse{Status: "ok"})
}
