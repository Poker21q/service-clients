package users_leaderboard_get

import (
	"context"
	"net/http"
	"strconv"

	"service-boilerplate-go/internal/pkg/response"
	"service-boilerplate-go/internal/services/users/entities"

	"service-boilerplate-go/internal/generated/api"
)

type Logger interface {
	Error(ctx context.Context, msg string)
}

type Service interface {
	GetUsersLeaderboard(ctx context.Context, limit, offset int) (entities.UsersLeaderboard, error)
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

	limit, offset := 10, 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v < 100 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	leaders, err := h.service.GetUsersLeaderboard(ctx, limit, offset)
	if err != nil {
		h.logger.Error(ctx, "failed to get leaderboard: "+err.Error())
		response.ErrorDomain(w, err)
		return
	}

	resp := make([]api.LeaderboardUser, len(leaders))
	for i, u := range leaders {
		resp[i] = api.LeaderboardUser{
			Id:       u.ID,
			Username: u.Username,
			Points:   u.Points,
		}
	}
	response.OkJSON(w, resp)
}
