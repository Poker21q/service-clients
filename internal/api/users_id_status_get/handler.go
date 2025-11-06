package users_id_status_get

import (
	"context"
	"net/http"

	"service-boilerplate-go/internal/generated/api"
	"service-boilerplate-go/internal/pkg/middlewares/jwtauth"
	"service-boilerplate-go/internal/pkg/response"
	"service-boilerplate-go/internal/services/users/entities"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Logger interface {
	Error(ctx context.Context, msg string)
}

type Service interface {
	GetUserStatus(ctx context.Context, userID uuid.UUID) (*entities.UserStatus, error)
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
		response.ErrorStatus(w, http.StatusUnauthorized)
		return
	}

	// получаем userID из пути
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

	// вызываем сервис
	statusEntity, err := h.service.GetUserStatus(ctx, userID)
	if err != nil {
		response.ErrorDomain(w, err)
		return
	}

	// мапим сущность в DTO
	statusDTO := mapUserStatusToDTO(statusEntity)

	// возвращаем ответ
	response.OkJSON(w, statusDTO)
}

// mapUserStatusToDTO конвертирует entities.UserStatus в api.UserStatus
func mapUserStatusToDTO(us *entities.UserStatus) *api.UserStatus {
	completed := make([]api.CompletedTask, len(us.CompletedTasks))
	for i, t := range us.CompletedTasks {
		var desc *string
		if t.Description != "" {
			desc = &t.Description
		}

		var meta *map[string]string
		if t.Metadata != nil && len(t.Metadata) > 0 {
			meta = &t.Metadata
		}

		completed[i] = api.CompletedTask{
			Id:          t.ID,
			Code:        t.Code,
			Description: desc,
			Points:      t.Points,
			Metadata:    meta,
			CompletedAt: t.CompletedAt,
		}
	}

	return &api.UserStatus{
		Id:             us.ID,
		Username:       us.Username,
		Points:         us.Points,
		ReferrerId:     us.ReferrerID,
		CompletedTasks: completed,
	}
}
