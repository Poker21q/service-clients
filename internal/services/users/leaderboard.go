package users

import (
	"context"

	"service-boilerplate-go/internal/services/users/entities"
)

// GetUsersLeaderboard возвращает топ пользователей по points
func (s *Service) GetUsersLeaderboard(ctx context.Context, limit, offset int) (entities.UsersLeaderboard, error) {
	return s.storage.GetUsersLeaderboard(ctx, limit, offset)
}
