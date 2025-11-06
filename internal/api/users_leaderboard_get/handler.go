package users_leaderboard_get

import (
	"context"
	"net/http"
	"strconv"

	"service-boilerplate-go/internal/generated/api"
	"service-boilerplate-go/internal/pkg/response"
	"service-boilerplate-go/internal/service/entities"
)

type Logger interface {
	Error(ctx context.Context, msg string)
	Warn(ctx context.Context, msg string)
	Info(ctx context.Context, msg string)

	WithFields(ctx context.Context, fields map[string]any) context.Context
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
		} else {
			ctx = h.logger.WithFields(ctx, map[string]any{
				"limit_param": l,
				"error":       err,
			})
			h.logger.Warn(ctx, "invalid limit parameter, using default 10")
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		} else {
			ctx = h.logger.WithFields(ctx, map[string]any{
				"offset_param": o,
				"error":        err,
			})
			h.logger.Warn(ctx, "invalid offset parameter, using default 0")
		}
	}

	ctx = h.logger.WithFields(ctx, map[string]any{
		"limit":  limit,
		"offset": offset,
	})

	leaders, err := h.service.GetUsersLeaderboard(ctx, limit, offset)
	if err != nil {
		ctx = h.logger.WithFields(ctx, map[string]any{
			"error": err.Error(),
		})
		h.logger.Error(ctx, "failed to get leaderboard")
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

	ctx = h.logger.WithFields(ctx, map[string]any{
		"returned_count": len(resp),
	})
	h.logger.Info(ctx, "users leaderboard retrieved successfully")

	response.OkJSON(w, resp)
}
