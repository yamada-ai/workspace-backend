package repository

import (
	"context"
	"time"

	"github.com/yamada-ai/workspace-backend/domain"
)

// SessionRepository defines the interface for session persistence operations
type SessionRepository interface {
	// Save creates or updates a session (upsert)
	// If session.ID == 0, creates a new session and sets the ID
	// Otherwise, updates the existing session
	Save(ctx context.Context, session *domain.Session) error

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

	// FindByUserIDAndDateRange retrieves sessions for a user within a date range
	FindByUserIDAndDateRange(ctx context.Context, userID int64, startTime, endTime time.Time) ([]*domain.Session, error)

	// FindActiveByUserIDWithTx retrieves the active session within a transaction
	FindActiveByUserIDWithTx(ctx context.Context, tx Tx, userID int64) (*domain.Session, error)

	// CreateWithTx creates a new session within a transaction
	CreateWithTx(ctx context.Context, tx Tx, session *domain.Session) error

	// FindAllActive retrieves all active sessions with user information
	FindAllActive(ctx context.Context) ([]domain.SessionInfo, error)
}
