package entities

import (
	"time"

	"github.com/google/uuid"
)

// Task - задача, которую можно выполнить для получения очков
type Task struct {
	ID           uuid.UUID
	Code         string
	Description  string
	RewardPoints int
	CreatedAt    time.Time
}
