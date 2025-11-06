package service

import (
	"context"

	entities2 "service-boilerplate-go/internal/service/entities"

	"github.com/google/uuid"
)

type Storage interface {
	CreateUser(ctx context.Context, username, passwordHash string) (*entities2.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*entities2.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entities2.User, error)

	GetUsersLeaderboard(ctx context.Context, limit, offset int) (entities2.UsersLeaderboard, error)

	IsUserExists(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateUserReferrer(ctx context.Context, userID, referrerID uuid.UUID) error

	IsTaskExists(ctx context.Context, taskID uuid.UUID) (bool, error)
	IsTaskCompleted(ctx context.Context, userID, taskID uuid.UUID) (bool, error)

	GetUserStatus(ctx context.Context, userID uuid.UUID) (*entities2.UserStatus, error)

	MarkTaskCompleted(ctx context.Context, userID, taskID uuid.UUID) (userTaskID uuid.UUID, err error)
	MarkTaskMetadata(ctx context.Context, userTaskID uuid.UUID, metadata map[string]string) error
}

type Service struct {
	storage   Storage
	jwtSecret []byte
}

func New(storage Storage, jwtSecret string) *Service {
	return &Service{
		storage:   storage,
		jwtSecret: []byte(jwtSecret),
	}
}
