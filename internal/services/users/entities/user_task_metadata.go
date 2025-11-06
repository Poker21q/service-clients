package entities

import "github.com/google/uuid"

// UserTaskMetadata - дополнительные данные для выполнения задачи
type UserTaskMetadata struct {
	ID         uuid.UUID
	UserTaskID uuid.UUID // ссылка на выполнение
	Key        string
	Value      string
}
