package storage

import (
	"context"

	"service-boilerplate-go/internal/services/users/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) GetUserStatus(ctx context.Context, userID uuid.UUID) (*entities.UserStatus, error) {
	// 1. Получаем данные пользователя
	const userQuery = `
		SELECT id, username, points, referrer_id
		FROM users
		WHERE id = $1
	`
	var status entities.UserStatus
	err := s.db.QueryRow(ctx, userQuery, userID).Scan(
		&status.ID,
		&status.Username,
		&status.Points,
		&status.ReferrerID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entities.ErrUserNotFound
		}
		return nil, err
	}

	// 2. Получаем все выполненные задачи
	tasks, err := s.GetCompletedTasks(ctx, userID)
	if err != nil {
		return nil, err
	}
	status.CompletedTasks = make([]entities.CompletedTask, len(tasks))
	for i, t := range tasks {
		status.CompletedTasks[i] = *t
	}

	return &status, nil
}

func (s *Storage) GetCompletedTasks(ctx context.Context, userID uuid.UUID) ([]*entities.CompletedTask, error) {
	tasks, err := s.fetchUserTasks(ctx, userID)
	if err != nil || len(tasks) == 0 {
		return tasks, err
	}

	metaMap, err := s.fetchTasksMetadata(ctx, tasks)
	if err != nil {
		return nil, err
	}

	s.attachMetadata(tasks, metaMap)
	return tasks, nil
}

func (s *Storage) fetchUserTasks(ctx context.Context, userID uuid.UUID) ([]*entities.CompletedTask, error) {
	const query = `
		SELECT ut.id, t.code, t.description, t.reward_points, ut.completed_at
		FROM user_tasks ut
		JOIN tasks t ON ut.task_id = t.id
		WHERE ut.user_id = $1
		ORDER BY ut.completed_at ASC
	`
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*entities.CompletedTask
	for rows.Next() {
		var t entities.CompletedTask
		if err := rows.Scan(&t.ID, &t.Code, &t.Description, &t.Points, &t.CompletedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}

	return tasks, nil
}

func (s *Storage) fetchTasksMetadata(ctx context.Context, tasks []*entities.CompletedTask) (map[uuid.UUID]map[string]string, error) {
	userTaskIDs := make([]uuid.UUID, 0, len(tasks))
	for _, t := range tasks {
		userTaskIDs = append(userTaskIDs, t.ID)
	}

	const query = `
		SELECT user_task_id, key, value
		FROM user_task_metadata
		WHERE user_task_id = ANY($1)
	`
	rows, err := s.db.Query(ctx, query, userTaskIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metaMap := make(map[uuid.UUID]map[string]string)
	for rows.Next() {
		var utID uuid.UUID
		var k, v string
		if err := rows.Scan(&utID, &k, &v); err != nil {
			return nil, err
		}
		if _, exists := metaMap[utID]; !exists {
			metaMap[utID] = make(map[string]string)
		}
		metaMap[utID][k] = v
	}

	return metaMap, nil
}

func (s *Storage) attachMetadata(tasks []*entities.CompletedTask, metaMap map[uuid.UUID]map[string]string) {
	for _, t := range tasks {
		if m, ok := metaMap[t.ID]; ok {
			t.Metadata = m
		} else {
			t.Metadata = map[string]string{}
		}
	}
}
