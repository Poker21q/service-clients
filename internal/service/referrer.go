package service

import (
	"context"

	entities2 "service-boilerplate-go/internal/service/entities"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func (s *Service) InputReferrer(ctx context.Context, userID, referrerID uuid.UUID) error {
	var (
		user           *entities2.User
		userExists     bool
		referrerExists bool
	)

	g, groupCtx := errgroup.WithContext(ctx)

	// проверяем пользователя и реферера параллельно
	g.Go(func() error {
		u, err := s.storage.GetUserByID(groupCtx, userID)
		if err != nil {
			return err
		}
		if u == nil {
			return entities2.ErrUserNotFound
		}
		user = u
		userExists = true
		return nil
	})

	g.Go(func() error {
		exists, err := s.storage.IsUserExists(groupCtx, referrerID)
		if err != nil {
			return err
		}
		if !exists {
			return entities2.ErrUserNotFound
		}
		referrerExists = true
		return nil
	})

	// ждём завершения проверок
	if err := g.Wait(); err != nil {
		return err
	}

	if !userExists || !referrerExists {
		return entities2.ErrUserNotFound
	}

	if user.ReferrerID != nil {
		return entities2.ErrReferrerAlreadySet
	}

	// обновляем поле referrer_id
	if err := s.storage.UpdateUserReferrer(ctx, userID, referrerID); err != nil {
		return err
	}

	return nil
}
