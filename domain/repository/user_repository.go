package repository

import (
	"context"

	"github.com/yamada-ai/workspace-backend/domain"
)

// Tx represents a database transaction
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// UserRepository defines the interface for user persistence operations
type UserRepository interface {
	// FindByName retrieves a user by name
	FindByName(ctx context.Context, name string) (*domain.User, error)

	// FindByID retrieves a user by ID
	FindByID(ctx context.Context, id int64) (*domain.User, error)

	// Save creates a new user or updates an existing one
	Save(ctx context.Context, user *domain.User) error

	// BeginTx starts a new transaction
	BeginTx(ctx context.Context) (Tx, error)

	// FindByNameWithTx retrieves a user by name within a transaction with row lock
	FindByNameWithTx(ctx context.Context, tx Tx, name string) (*domain.User, error)

	// SaveWithTx creates a new user within a transaction
	SaveWithTx(ctx context.Context, tx Tx, user *domain.User) error
}
