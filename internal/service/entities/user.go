package entities

import (
	"time"

	"github.com/google/uuid"
)

// User - пользователь системы
type User struct {
	ID         uuid.UUID
	Username   string
	Password   string
	Points     int
	ReferrerID *uuid.UUID // может быть nil
	CreatedAt  time.Time
}
