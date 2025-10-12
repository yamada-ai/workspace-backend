package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yamada-ai/workspace-backend/domain"
	domainRepo "github.com/yamada-ai/workspace-backend/domain/repository"
	"github.com/yamada-ai/workspace-backend/infrastructure/database/sqlc"
)

// Ensure userRepositoryImpl implements domain.UserRepository
var _ domainRepo.UserRepository = (*userRepositoryImpl)(nil)

type userRepositoryImpl struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

// txWrapper wraps pgx.Tx to implement domainRepo.Tx
type txWrapper struct {
	tx pgx.Tx
}

func (tw *txWrapper) Commit(ctx context.Context) error {
	return tw.tx.Commit(ctx)
}

func (tw *txWrapper) Rollback(ctx context.Context) error {
	return tw.tx.Rollback(ctx)
}

// NewUserRepository creates a new user repository implementation
func NewUserRepository(queries *sqlc.Queries) domainRepo.UserRepository {
	return &userRepositoryImpl{
		queries: queries,
		pool:    nil, // Will be set when needed
	}
}

// NewUserRepositoryWithPool creates a new user repository with pool support
func NewUserRepositoryWithPool(pool *pgxpool.Pool) domainRepo.UserRepository {
	return &userRepositoryImpl{
		queries: sqlc.New(pool),
		pool:    pool,
	}
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

func (r *userRepositoryImpl) BeginTx(ctx context.Context) (domainRepo.Tx, error) {
	if r.pool == nil {
		return nil, errors.New("pool not available for transactions")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &txWrapper{tx: tx}, nil
}

func (r *userRepositoryImpl) FindByNameWithTx(ctx context.Context, tx domainRepo.Tx, name string) (*domain.User, error) {
	wrapper, ok := tx.(*txWrapper)
	if !ok {
		return nil, errors.New("invalid transaction type")
	}

	queries := sqlc.New(wrapper.tx)
	user, err := queries.FindUserByNameForUpdate(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUser(user), nil
}

func (r *userRepositoryImpl) SaveWithTx(ctx context.Context, tx domainRepo.Tx, user *domain.User) error {
	wrapper, ok := tx.(*txWrapper)
	if !ok {
		return errors.New("invalid transaction type")
	}

	queries := sqlc.New(wrapper.tx)
	created, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		Name:      user.Name,
		Tier:      int32(user.Tier.Int()),
		CreatedAt: pgtype.Timestamp{Time: user.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: user.UpdatedAt, Valid: true},
	})
	if err != nil {
		// Check for duplicate key error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUserAlreadyExists
		}
		return err
	}
	user.ID = int64(created.ID)
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
