package entities

import (
	"time"

	"github.com/google/uuid"
)

type CompletedTask struct {
	ID          uuid.UUID
	Code        string
	Description string
	Points      int
	Metadata    map[string]string
	CompletedAt time.Time
}

type UserStatus struct {
	ID             uuid.UUID
	Username       string
	Points         int
	ReferrerID     *uuid.UUID
	CompletedTasks []CompletedTask
}
