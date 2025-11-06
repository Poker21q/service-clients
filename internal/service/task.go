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
