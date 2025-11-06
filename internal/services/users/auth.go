package users

import (
	"context"
	"errors"

	"service-boilerplate-go/internal/services/users/entities"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Auth(ctx context.Context, username, password string) (token string, userID uuid.UUID, err error) {
	user, err := s.storage.GetUserByUsername(ctx, username)
	if err != nil && !errors.Is(err, entities.ErrUserNotFound) {
		return "", uuid.Nil, err
	}

	if user == nil {
		hash, err := hashPassword(password)
		if err != nil {
			return "", uuid.Nil, err
		}
		user, err = s.storage.CreateUser(ctx, username, hash)
		if err != nil {
			return "", uuid.Nil, err
		}
	} else {
		if err := checkPasswordHash(password, user.Password); err != nil {
			return "", uuid.Nil, entities.ErrInvalidCredentials
		}
	}

	token, err = s.generateToken(user.ID)
	if err != nil {
		return "", uuid.Nil, err
	}

	userID = user.ID
	return token, userID, err
}

// hashPassword создаёт bcrypt-хэш
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPasswordHash проверяет пароль на соответствие хэшу
func checkPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// generateToken создаёт JWT без срока жизни
func (s *Service) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.jwtSecret)
}
