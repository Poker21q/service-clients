package storage

import (
	"context"

	"service-boilerplate-go/internal/services/users/entities"

	"github.com/google/uuid"
)

func (s *Storage) UpdateUserReferrer(ctx context.Context, userID, referrerID uuid.UUID) error {
	const query = `
		UPDATE users
		SET referrer_id = $1
		WHERE id = $2
	`

	tag, err := s.db.Exec(ctx, query, referrerID, userID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		// Если обновление не затронуло ни одной строки — пользователя нет
		return entities.ErrUserNotFound
	}

	return nil
}
