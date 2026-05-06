package apperrors

import "errors"

// Sentinel errors for business-level error handling.
// Services wrap these with context using fmt.Errorf("context: %w", ErrXxx).
// Handlers check with errors.Is(err, apperrors.ErrXxx) to determine HTTP status codes.

// Generic resource errors
var (
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("conflict")
	ErrForbidden = errors.New("forbidden")
	ErrInUse     = errors.New("in use")
)

// Auth errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrEmailTaken         = errors.New("email already taken")
)

// Validation errors
var (
	ErrValidation = errors.New("validation error")
)

// File errors
var (
	ErrFileTooLarge    = errors.New("file too large")
	ErrInvalidFileType = errors.New("invalid file type")
	ErrNoPhoto         = errors.New("no photo")
)

// Swap errors
var (
	ErrSelfAction     = errors.New("self action not allowed")
	ErrDuplicate      = errors.New("duplicate")
	ErrWrongStatus    = errors.New("wrong status")
	ErrNotParticipant = errors.New("not a participant")
)