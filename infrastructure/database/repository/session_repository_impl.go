package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/yamada-ai/workspace-backend/domain"
	domainRepo "github.com/yamada-ai/workspace-backend/domain/repository"
	"github.com/yamada-ai/workspace-backend/infrastructure/database/sqlc"
)

// Ensure sessionRepositoryImpl implements domain.SessionRepository
var _ domainRepo.SessionRepository = (*sessionRepositoryImpl)(nil)

type sessionRepositoryImpl struct {
	queries *sqlc.Queries
}

// NewSessionRepository creates a new session repository implementation
func NewSessionRepository(queries *sqlc.Queries) domainRepo.SessionRepository {
	return &sessionRepositoryImpl{queries: queries}
}

func (r *sessionRepositoryImpl) FindActiveByUserIDWithTx(ctx context.Context, tx domainRepo.Tx, userID int64) (*domain.Session, error) {
	wrapper, ok := tx.(*txWrapper)
	if !ok {
		return nil, errors.New("invalid transaction type")
	}

	queries := sqlc.New(wrapper.tx)
	session, err := queries.FindActiveSessionByUserID(ctx, int32(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}
	return toDomainSession(session), nil
}

func (r *sessionRepositoryImpl) CreateWithTx(ctx context.Context, tx domainRepo.Tx, session *domain.Session) error {
	wrapper, ok := tx.(*txWrapper)
	if !ok {
		return errors.New("invalid transaction type")
	}

	queries := sqlc.New(wrapper.tx)

	var iconID pgtype.Int4
	if session.IconID != nil {
		iconID = pgtype.Int4{Int32: int32(*session.IconID), Valid: true}
	}

	var workName pgtype.Text
	if session.WorkName != "" {
		workName = pgtype.Text{String: session.WorkName, Valid: true}
	}

	created, err := queries.CreateSession(ctx, sqlc.CreateSessionParams{
		UserID:     int32(session.UserID),
		WorkName:   workName,
		StartTime:  pgtype.Timestamp{Time: session.StartTime, Valid: true},
		PlannedEnd: pgtype.Timestamp{Time: session.PlannedEnd, Valid: true},
		IconID:     iconID,
		CreatedAt:  pgtype.Timestamp{Time: session.CreatedAt, Valid: true},
		UpdatedAt:  pgtype.Timestamp{Time: session.UpdatedAt, Valid: true},
	})
	if err != nil {
		return err
	}

	session.ID = int64(created.ID)
	return nil
}

// Save creates or updates a session
func (r *sessionRepositoryImpl) Save(ctx context.Context, session *domain.Session) error {
	if session.ID == 0 {
		// Create new session
		return r.Create(ctx, session)
	}
	// Update existing session
	return r.Update(ctx, session)
}

func (r *sessionRepositoryImpl) Create(ctx context.Context, session *domain.Session) error {
	var iconID pgtype.Int4
	if session.IconID != nil {
		iconID = pgtype.Int4{Int32: int32(*session.IconID), Valid: true}
	}

	var workName pgtype.Text
	if session.WorkName != "" {
		workName = pgtype.Text{String: session.WorkName, Valid: true}
	}

	created, err := r.queries.CreateSession(ctx, sqlc.CreateSessionParams{
		UserID:     int32(session.UserID),
		WorkName:   workName,
		StartTime:  pgtype.Timestamp{Time: session.StartTime, Valid: true},
		PlannedEnd: pgtype.Timestamp{Time: session.PlannedEnd, Valid: true},
		IconID:     iconID,
		CreatedAt:  pgtype.Timestamp{Time: session.CreatedAt, Valid: true},
		UpdatedAt:  pgtype.Timestamp{Time: session.UpdatedAt, Valid: true},
	})
	if err != nil {
		return err
	}

	session.ID = int64(created.ID)
	return nil
}

func (r *sessionRepositoryImpl) FindByID(ctx context.Context, id int64) (*domain.Session, error) {
	session, err := r.queries.FindSessionByID(ctx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}
	return toDomainSession(session), nil
}

func (r *sessionRepositoryImpl) FindActiveByUserID(ctx context.Context, userID int64) (*domain.Session, error) {
	session, err := r.queries.FindActiveSessionByUserID(ctx, int32(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}
	return toDomainSession(session), nil
}

func (r *sessionRepositoryImpl) Update(ctx context.Context, session *domain.Session) error {
	// Update planned_end if session is being extended
	if session.IsActive() {
		_, err := r.queries.UpdateSessionPlannedEnd(ctx, sqlc.UpdateSessionPlannedEndParams{
			ID:         int32(session.ID),
			PlannedEnd: pgtype.Timestamp{Time: session.PlannedEnd, Valid: true},
			UpdatedAt:  pgtype.Timestamp{Time: session.UpdatedAt, Valid: true},
		})
		return err
	}

	// Complete session if actual_end is set
	if session.ActualEnd != nil {
		_, err := r.queries.CompleteSession(ctx, sqlc.CompleteSessionParams{
			ID:        int32(session.ID),
			ActualEnd: pgtype.Timestamp{Time: *session.ActualEnd, Valid: true},
			UpdatedAt: pgtype.Timestamp{Time: session.UpdatedAt, Valid: true},
		})
		return err
	}

	return nil
}

func (r *sessionRepositoryImpl) ListByUserID(ctx context.Context, userID int64, limit, offset int32) ([]*domain.Session, error) {
	sessions, err := r.queries.ListUserSessions(ctx, sqlc.ListUserSessionsParams{
		UserID: int32(userID),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Session, len(sessions))
	for i, s := range sessions {
		result[i] = toDomainSession(s)
	}
	return result, nil
}

// toDomainSession converts sqlc.Session to domain.Session
func toDomainSession(session sqlc.Session) *domain.Session {
	var workName string
	if session.WorkName.Valid {
		workName = session.WorkName.String
	}

	var actualEnd *time.Time
	if session.ActualEnd.Valid {
		t := session.ActualEnd.Time
		actualEnd = &t
	}

	var iconID *int64
	if session.IconID.Valid {
		id := int64(session.IconID.Int32)
		iconID = &id
	}

	return &domain.Session{
		ID:         int64(session.ID),
		UserID:     int64(session.UserID),
		WorkName:   workName,
		StartTime:  session.StartTime.Time,
		PlannedEnd: session.PlannedEnd.Time,
		ActualEnd:  actualEnd,
		IconID:     iconID,
		CreatedAt:  session.CreatedAt.Time,
		UpdatedAt:  session.UpdatedAt.Time,
	}
}

// FindAllActive retrieves all active sessions with user information
func (r *sessionRepositoryImpl) FindAllActive(ctx context.Context) ([]domain.SessionInfo, error) {
	rows, err := r.queries.GetActiveSessions(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.SessionInfo, 0, len(rows))
	for _, row := range rows {
		var iconID *int64
		if row.IconID.Valid {
			id := int64(row.IconID.Int32)
			iconID = &id
		}

		result = append(result, domain.SessionInfo{
			SessionID:  int64(row.ID),
			UserID:     int64(row.UserID),
			UserName:   row.UserName,
			WorkName:   row.WorkName.String,
			Tier:       int(row.UserTier),
			IconID:     iconID,
			StartTime:  row.StartTime.Time,
			PlannedEnd: row.PlannedEnd.Time,
		})
	}

	return result, nil
}
