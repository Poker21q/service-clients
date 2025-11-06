package entities

import (
	"time"

	"github.com/google/uuid"
)

// UserTask - выполнение задачи конкретным пользователем
type UserTask struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	TaskID      uuid.UUID
	CompletedAt time.Time
}
