package entities

import "github.com/google/uuid"

type UserLeader struct {
	ID       uuid.UUID
	Username string
	Points   int
}

type UsersLeaderboard []UserLeader
