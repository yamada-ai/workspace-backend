package repository

import (
	"context"

	"github.com/yamada-ai/workspace-backend/domain"
)

// SessionRepository defines the interface for session persistence operations
type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, session *domain.Session) error

	// FindByID retrieves a session by ID
	FindByID(ctx context.Context, id int64) (*domain.Session, error)

	// FindActiveByUserID retrieves the active session for a user (actual_end is NULL)
	FindActiveByUserID(ctx context.Context, userID int64) (*domain.Session, error)

	// Update updates an existing session
	Update(ctx context.Context, session *domain.Session) error

	// ListByUserID retrieves sessions for a user with pagination
	ListByUserID(ctx context.Context, userID int64, limit, offset int32) ([]*domain.Session, error)
}
