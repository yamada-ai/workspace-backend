package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yamada-ai/workspace-backend/domain"
	domainRepo "github.com/yamada-ai/workspace-backend/domain/repository"
	"github.com/yamada-ai/workspace-backend/infrastructure/database/sqlc"
)

// Ensure userRepositoryImpl implements domain.UserRepository
var _ domainRepo.UserRepository = (*userRepositoryImpl)(nil)

type userRepositoryImpl struct {
	queries *sqlc.Queries
}

// NewUserRepository creates a new user repository implementation
func NewUserRepository(queries *sqlc.Queries) domainRepo.UserRepository {
	return &userRepositoryImpl{queries: queries}
}

func (r *userRepositoryImpl) FindByName(ctx context.Context, name string) (*domain.User, error) {
	user, err := r.queries.FindUserByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUser(user), nil
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := r.queries.FindUserByID(ctx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUser(user), nil
}

func (r *userRepositoryImpl) Save(ctx context.Context, user *domain.User) error {
	if user.ID == 0 {
		// Create new user
		created, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
			Name:      user.Name,
			Tier:      int32(user.Tier.Int()),
			CreatedAt: pgtype.Timestamp{Time: user.CreatedAt, Valid: true},
			UpdatedAt: pgtype.Timestamp{Time: user.UpdatedAt, Valid: true},
		})
		if err != nil {
			return err
		}
		user.ID = int64(created.ID)
		return nil
	}

	// Update existing user
	updated, err := r.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:        int32(user.ID),
		Tier:      int32(user.Tier.Int()),
		UpdatedAt: pgtype.Timestamp{Time: user.UpdatedAt, Valid: true},
	})
	if err != nil {
		return err
	}

	*user = *toDomainUser(updated)
	return nil
}

// toDomainUser converts sqlc.User to domain.User
func toDomainUser(user sqlc.User) *domain.User {
	// Convert int32 tier to domain.Tier
	tier, err := domain.ParseTier(strconv.Itoa(int(user.Tier)))
	if err != nil || tier == domain.TierUnknown {
		// Fallback: treat as Tier1 if parsing fails
		tier = domain.Tier1
	}

	return &domain.User{
		ID:        int64(user.ID),
		Name:      user.Name,
		Tier:      tier,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}
}
