package service

import (
	"context"

	"service-boilerplate-go/internal/service/entities"

	"github.com/google/uuid"
)

func (s *Service) GetUserStatus(ctx context.Context, userID uuid.UUID) (*entities.UserStatus, error) {
	return s.storage.GetUserStatus(ctx, userID)
}
