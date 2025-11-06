package storage

import (
	"context"
	"errors"
	"time"

	"service-boilerplate-go/internal/service/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TaskModel — структура для работы с таблицей tasks
type TaskModel struct {
	ID           uuid.UUID
	Code         string
	Description  string
	RewardPoints int
	CreatedAt    time.Time
}

// UserTaskModel — структура для таблицы user_tasks
type UserTaskModel struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	TaskID      uuid.UUID
	CompletedAt time.Time
}

// Проверяем, существует ли задача
func (s *Storage) IsTaskExists(ctx context.Context, taskID uuid.UUID) (bool, error) {
	const query = `SELECT 1 FROM tasks WHERE id = $1`
	var tmp int
	err := s.db.QueryRow(ctx, query, taskID).Scan(&tmp)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Проверяем, выполнена ли задача пользователем
func (s *Storage) IsTaskCompleted(ctx context.Context, userID, taskID uuid.UUID) (bool, error) {
	const query = `SELECT 1 FROM user_tasks WHERE user_id = $1 AND task_id = $2`
	var tmp int
	err := s.db.QueryRow(ctx, query, userID, taskID).Scan(&tmp)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// MarkTaskCompleted вставляет запись о выполненной задаче и обновляет баллы пользователя.
// Возвращает ID созданной записи в user_tasks.
func (s *Storage) MarkTaskCompleted(ctx context.Context, userID, taskID uuid.UUID) (uuid.UUID, error) {
	const insertQuery = `
		INSERT INTO user_tasks (user_id, task_id, completed_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var userTaskID uuid.UUID
	err := s.db.QueryRow(ctx, insertQuery, userID, taskID, time.Now()).Scan(&userTaskID)
	if err != nil {
		return uuid.Nil, err
	}

	// обновление баллов пользователя
	const updateQuery = `
		UPDATE users 
		SET points = points + (SELECT reward_points FROM tasks WHERE id = $1)
		WHERE id = $2
	`
	_, err = s.db.Exec(ctx, updateQuery, taskID, userID)
	if err != nil {
		return uuid.Nil, err
	}

	return userTaskID, nil
}

// MarkTaskMetadata батчево сохраняет метадату для выполненного задания
// и возвращает доменную ошибку при любой ошибке вставки
func (s *Storage) MarkTaskMetadata(ctx context.Context, userTaskID uuid.UUID, metadata map[string]string) error {
	if len(metadata) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	const query = `
		INSERT INTO user_task_metadata (user_task_id, key, value)
		VALUES ($1, $2, $3)
	`

	for key, value := range metadata {
		batch.Queue(query, userTaskID, key, value)
	}

	br := s.db.SendBatch(ctx, batch)
	defer func() {
		if err := br.Close(); err != nil {
			ctx = context.WithoutCancel(ctx)
			s.logger.Error(ctx, err.Error())
		}
	}()

	for range metadata {
		if _, err := br.Exec(); err != nil {
			return entities.ErrTaskMetadataAlreadyExists
		}
	}

	return nil
}
