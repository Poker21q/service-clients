package storage

import (
	"context"
	"errors"
	"time"

	"service-boilerplate-go/internal/service/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// UserModel — структура для работы с таблицей service
type UserModel struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	Points       int
	ReferrerID   *uuid.UUID
	CreatedAt    time.Time
}

// CreateUser создаёт нового пользователя и возвращает сущность
func (s *Storage) CreateUser(ctx context.Context, username, passwordHash string) (*entities.User, error) {
	const query = `
		INSERT INTO users (id, username, password_hash, points, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, password_hash, points, referrer_id, created_at
	`

	var m UserModel
	err := s.db.QueryRow(ctx, query, uuid.New(), username, passwordHash, int64(0), time.Now()).Scan(
		&m.ID,
		&m.Username,
		&m.PasswordHash,
		&m.Points,
		&m.ReferrerID,
		&m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return mapUserModelToEntity(&m), nil
}

// GetUserByUsername возвращает сущность пользователя по username
func (s *Storage) GetUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	const query = `
		SELECT id, username, password_hash, points, referrer_id, created_at
		FROM users
		WHERE username = $1
	`

	var m UserModel
	err := s.db.QueryRow(ctx, query, username).Scan(
		&m.ID,
		&m.Username,
		&m.PasswordHash,
		&m.Points,
		&m.ReferrerID,
		&m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrUserNotFound
		}
		return nil, err
	}

	return mapUserModelToEntity(&m), nil
}

func (s *Storage) IsUserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	const query = `SELECT 1 FROM users WHERE id = $1`
	var tmp int
	err := s.db.QueryRow(ctx, query, id).Scan(&tmp)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetUserByID возвращает сущность пользователя по UUID
func (s *Storage) GetUserByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	const query = `
		SELECT id, username, password_hash, points, referrer_id, created_at
		FROM users
		WHERE id = $1
	`

	var m UserModel
	err := s.db.QueryRow(ctx, query, id).Scan(
		&m.ID,
		&m.Username,
		&m.PasswordHash,
		&m.Points,
		&m.ReferrerID,
		&m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrUserNotFound
		}
		return nil, err
	}

	return mapUserModelToEntity(&m), nil
}

// mapUserModelToEntity конвертирует модель базы в сущность
func mapUserModelToEntity(m *UserModel) *entities.User {
	return &entities.User{
		ID:         m.ID,
		Username:   m.Username,
		Password:   m.PasswordHash,
		Points:     m.Points,
		ReferrerID: m.ReferrerID,
		CreatedAt:  m.CreatedAt,
	}
}

type UserLeaderBoardRow struct {
	ID       uuid.UUID
	Username string
	Points   int
}

// GetUsersLeaderboard возвращает топ пользователей по points с пагинацией
func (s *Storage) GetUsersLeaderboard(ctx context.Context, limit, offset int) (entities.UsersLeaderboard, error) {
	const query = `
		SELECT id, username, points
		FROM users
		ORDER BY points DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return entities.UsersLeaderboard{}, err
	}
	defer rows.Close()

	// сканируем в модель БД
	var models []UserModel
	for rows.Next() {
		var m UserModel
		if err := rows.Scan(&m.ID, &m.Username, &m.Points); err != nil {
			return entities.UsersLeaderboard{}, err
		}
		models = append(models, m)
	}

	// мапим в сущности
	leaders := make([]entities.UserLeader, len(models))
	for i, m := range models {
		leaders[i] = entities.UserLeader{
			ID:       m.ID,
			Username: m.Username,
			Points:   m.Points,
		}
	}

	return leaders, nil
}

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
		if errors.Is(err, pgx.ErrNoRows) {
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
		ORDER BY ut.completed_at
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
