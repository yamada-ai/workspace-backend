package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yamada-ai/workspace-backend/domain"
	"github.com/yamada-ai/workspace-backend/domain/repository"
)

func TestGetUserInfo_Success(t *testing.T) {
	now := time.Date(2025, 11, 24, 15, 30, 0, 0, time.UTC)

	existingUser := &domain.User{
		ID:        42,
		Name:      "yamada",
		Tier:      domain.Tier1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Active session: started 30 minutes ago, planned to end in 45 minutes
	activeSession := &domain.Session{
		ID:         99,
		UserID:     42,
		WorkName:   "論文執筆",
		StartTime:  now.Add(-30 * time.Minute),
		PlannedEnd: now.Add(45 * time.Minute),
		ActualEnd:  nil,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Today's completed session: 120 minutes
	todaySession1 := &domain.Session{
		ID:         100,
		UserID:     42,
		WorkName:   "資格勉強",
		StartTime:  now.Add(-8 * time.Hour),
		PlannedEnd: now.Add(-6 * time.Hour),
		ActualEnd:  ptrTime(now.Add(-6 * time.Hour)),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// All sessions for lifetime calculation
	allSessions := []*domain.Session{
		activeSession,
		todaySession1,
		{
			ID:         101,
			UserID:     42,
			WorkName:   "開発",
			StartTime:  now.Add(-48 * time.Hour),
			PlannedEnd: now.Add(-46 * time.Hour),
			ActualEnd:  ptrTime(now.Add(-46 * time.Hour)),
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	userRepo := &mockUserRepository{
		findByNameFn: func(ctx context.Context, name string) (*domain.User, error) {
			if name == "yamada" {
				return existingUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}

	sessionRepo := &mockSessionRepository{
		findActiveByUserIDFn: func(ctx context.Context, userID int64) (*domain.Session, error) {
			if userID == 42 {
				return activeSession, nil
			}
			return nil, domain.ErrSessionNotFound
		},
		findByUserIDAndDateRangeFn: func(ctx context.Context, userID int64, startTime, endTime time.Time) ([]*domain.Session, error) {
			// Return today's sessions
			return []*domain.Session{activeSession, todaySession1}, nil
		},
		listByUserIDFn: func(ctx context.Context, userID int64, limit, offset int32) ([]*domain.Session, error) {
			// Return all sessions
			return allSessions, nil
		},
	}

	uc := NewGetUserInfoUseCase(userRepo, sessionRepo)
	uc.now = func() time.Time { return now }

	input := GetUserInfoInput{
		UserName: "yamada",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Remaining minutes: 45 minutes until planned_end
	if output.RemainingMinutes != 45 {
		t.Errorf("expected RemainingMinutes to be 45, got %d", output.RemainingMinutes)
	}

	// Today's total: active session (30 min) + completed session (120 min) = 150 min
	if output.TodayTotalMinutes != 150 {
		t.Errorf("expected TodayTotalMinutes to be 150, got %d", output.TodayTotalMinutes)
	}

	// Lifetime total: active (30) + todaySession1 (120) + old session (120) = 270 min
	if output.LifetimeTotalMinutes != 270 {
		t.Errorf("expected LifetimeTotalMinutes to be 270, got %d", output.LifetimeTotalMinutes)
	}

	if output.UserID != 42 {
		t.Errorf("expected UserID to be 42, got %d", output.UserID)
	}
}

func TestGetUserInfo_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepository{
		findByNameFn: func(ctx context.Context, name string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	sessionRepo := &mockSessionRepository{}

	uc := NewGetUserInfoUseCase(userRepo, sessionRepo)

	input := GetUserInfoInput{
		UserName: "nonexistent",
	}

	output, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}

	if output != nil {
		t.Errorf("expected output to be nil when error occurs, got %+v", output)
	}
}

func TestGetUserInfo_NoActiveSession(t *testing.T) {
	now := time.Now()

	existingUser := &domain.User{
		ID:        42,
		Name:      "yamada",
		Tier:      domain.Tier1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	userRepo := &mockUserRepository{
		findByNameFn: func(ctx context.Context, name string) (*domain.User, error) {
			return existingUser, nil
		},
	}

	sessionRepo := &mockSessionRepository{
		findActiveByUserIDFn: func(ctx context.Context, userID int64) (*domain.Session, error) {
			return nil, domain.ErrSessionNotFound
		},
	}

	uc := NewGetUserInfoUseCase(userRepo, sessionRepo)

	input := GetUserInfoInput{
		UserName: "yamada",
	}

	output, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}

	if output != nil {
		t.Errorf("expected output to be nil when error occurs, got %+v", output)
	}
}

// Helper function
func ptrTime(t time.Time) *time.Time {
	return &t
}

// Mock implementations
type mockUserRepository struct{
	findByNameFn func(ctx context.Context, name string) (*domain.User, error)
}

func (m *mockUserRepository) FindByName(ctx context.Context, name string) (*domain.User, error) {
	if m.findByNameFn != nil {
		return m.findByNameFn(ctx, name)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) Save(ctx context.Context, user *domain.User) error {
	return nil
}

func (m *mockUserRepository) BeginTx(ctx context.Context) (repository.Tx, error) {
	return nil, nil
}

func (m *mockUserRepository) FindByNameWithTx(ctx context.Context, tx repository.Tx, name string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) SaveWithTx(ctx context.Context, tx repository.Tx, user *domain.User) error {
	return nil
}

type mockSessionRepository struct {
	findActiveByUserIDFn       func(ctx context.Context, userID int64) (*domain.Session, error)
	findByUserIDAndDateRangeFn func(ctx context.Context, userID int64, startTime, endTime time.Time) ([]*domain.Session, error)
	listByUserIDFn             func(ctx context.Context, userID int64, limit, offset int32) ([]*domain.Session, error)
}

func (m *mockSessionRepository) Save(ctx context.Context, session *domain.Session) error {
	return nil
}

func (m *mockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	return nil
}

func (m *mockSessionRepository) FindByID(ctx context.Context, id int64) (*domain.Session, error) {
	return nil, domain.ErrSessionNotFound
}

func (m *mockSessionRepository) FindActiveByUserID(ctx context.Context, userID int64) (*domain.Session, error) {
	if m.findActiveByUserIDFn != nil {
		return m.findActiveByUserIDFn(ctx, userID)
	}
	return nil, domain.ErrSessionNotFound
}

func (m *mockSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	return nil
}

func (m *mockSessionRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int32) ([]*domain.Session, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID, limit, offset)
	}
	return nil, nil
}

func (m *mockSessionRepository) FindByUserIDAndDateRange(ctx context.Context, userID int64, startTime, endTime time.Time) ([]*domain.Session, error) {
	if m.findByUserIDAndDateRangeFn != nil {
		return m.findByUserIDAndDateRangeFn(ctx, userID, startTime, endTime)
	}
	return nil, nil
}

func (m *mockSessionRepository) FindActiveByUserIDWithTx(ctx context.Context, tx repository.Tx, userID int64) (*domain.Session, error) {
	return nil, domain.ErrSessionNotFound
}

func (m *mockSessionRepository) CreateWithTx(ctx context.Context, tx repository.Tx, session *domain.Session) error {
	return nil
}

func (m *mockSessionRepository) FindAllActive(ctx context.Context) ([]domain.SessionInfo, error) {
	return nil, nil
}
