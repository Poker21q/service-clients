package entities

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrReferrerAlreadySet = errors.New("referrer already set")

	ErrTaskNotFound              = errors.New("task not found")
	ErrTaskAlreadyCompleted      = errors.New("task already completed")
	ErrTaskMetadataAlreadyExists = errors.New("task metadata already exists")
)
