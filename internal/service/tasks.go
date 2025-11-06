package service

import (
	"context"

	"service-boilerplate-go/internal/service/entities"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func (s *Service) CompleteTask(ctx context.Context, userID, taskID uuid.UUID, metadata map[string]string) error {
	var (
		user       *entities.User
		taskExists bool
	)

	g, groupCtx := errgroup.WithContext(ctx)

	// проверяем пользователя и существование задания параллельно
	g.Go(func() error {
		u, err := s.storage.GetUserByID(groupCtx, userID)
		if err != nil {
			return err
		}
		if u == nil {
			return entities.ErrUserNotFound
		}
		user = u
		return nil
	})

	g.Go(func() error {
		exists, err := s.storage.IsTaskExists(groupCtx, taskID)
		if err != nil {
			return err
		}
		if !exists {
			return entities.ErrTaskNotFound
		}
		taskExists = true
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if user == nil {
		return entities.ErrUserNotFound
	}
	if !taskExists {
		return entities.ErrTaskNotFound
	}

	// проверяем, не было ли задание уже выполнено
	completed, err := s.storage.IsTaskCompleted(ctx, userID, taskID)
	if err != nil {
		return err
	}
	if completed {
		return entities.ErrTaskAlreadyCompleted
	}

	// сохраняем выполненное задание
	userTaskID, err := s.storage.MarkTaskCompleted(ctx, userID, taskID)
	if err != nil {
		return err
	}

	// сохраняем метадату батчем, если она есть
	if len(metadata) > 0 {
		if err := s.storage.MarkTaskMetadata(ctx, userTaskID, metadata); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) InputReferrer(ctx context.Context, userID, referrerID uuid.UUID) error {
	var (
		user           *entities.User
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
			return entities.ErrUserNotFound
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
			return entities.ErrUserNotFound
		}
		referrerExists = true
		return nil
	})

	// ждём завершения проверок
	if err := g.Wait(); err != nil {
		return err
	}

	if !userExists || !referrerExists {
		return entities.ErrUserNotFound
	}

	if user.ReferrerID != nil {
		return entities.ErrReferrerAlreadySet
	}

	// обновляем поле referrer_id
	if err := s.storage.UpdateUserReferrer(ctx, userID, referrerID); err != nil {
		return err
	}

	return nil
}
