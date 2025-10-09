package repository

import (
	"context"

	"github.com/yamada-ai/workspace-backend/domain"
)

// UserRepository defines the interface for user persistence operations
type UserRepository interface {
	// FindByName retrieves a user by name
	FindByName(ctx context.Context, name string) (*domain.User, error)

	// FindByID retrieves a user by ID
	FindByID(ctx context.Context, id int64) (*domain.User, error)

	// Save creates a new user or updates an existing one
	Save(ctx context.Context, user *domain.User) error
}
