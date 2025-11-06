package users_id_task_complete_post

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

	var req api.TaskCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	if req.TaskId == uuid.Nil {
		response.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	// используем метадату, если есть
	var metadata map[string]string
	if req.Metadata != nil {
		metadata = *req.Metadata
	}

	if err := h.service.CompleteTask(ctx, userID, req.TaskId, metadata); err != nil {
		response.ErrorDomain(w, err)
		return
	}

	response.OkJSON(w, api.TaskCompleteResponse{Status: "ok"})
}
