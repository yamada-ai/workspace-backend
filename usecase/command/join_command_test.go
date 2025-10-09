package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yamada-ai/workspace-backend/domain"
)

// Mock UserRepository
type mockUserRepository struct {
	findByNameFn func(ctx context.Context, name string) (*domain.User, error)
	findByIDFn   func(ctx context.Context, id int64) (*domain.User, error)
	saveFn       func(ctx context.Context, user *domain.User) error
}

func (m *mockUserRepository) FindByName(ctx context.Context, name string) (*domain.User, error) {
	if m.findByNameFn != nil {
		return m.findByNameFn(ctx, name)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) Save(ctx context.Context, user *domain.User) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, user)
	}
	// Simulate ID assignment
	if user.ID == 0 {
		user.ID = 1
	}
	return nil
}

// Mock SessionRepository
type mockSessionRepository struct {
	saveFn               func(ctx context.Context, session *domain.Session) error
	createFn             func(ctx context.Context, session *domain.Session) error
	findByIDFn           func(ctx context.Context, id int64) (*domain.Session, error)
	findActiveByUserIDFn func(ctx context.Context, userID int64) (*domain.Session, error)
	updateFn             func(ctx context.Context, session *domain.Session) error
	listByUserIDFn       func(ctx context.Context, userID int64, limit, offset int32) ([]*domain.Session, error)
}

func (m *mockSessionRepository) Save(ctx context.Context, session *domain.Session) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, session)
	}
	// Simulate ID assignment
	if session.ID == 0 {
		session.ID = 100
	}
	return nil
}

func (m *mockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if m.createFn != nil {
		return m.createFn(ctx, session)
	}
	// Simulate ID assignment
	if session.ID == 0 {
		session.ID = 100
	}
	return nil
}

func (m *mockSessionRepository) FindByID(ctx context.Context, id int64) (*domain.Session, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, errors.New("session not found")
}

func (m *mockSessionRepository) FindActiveByUserID(ctx context.Context, userID int64) (*domain.Session, error) {
	if m.findActiveByUserIDFn != nil {
		return m.findActiveByUserIDFn(ctx, userID)
	}
	return nil, errors.New("session not found")
}

func (m *mockSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, session)
	}
	return nil
}

func (m *mockSessionRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int32) ([]*domain.Session, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID, limit, offset)
	}
	return nil, nil
}

// Tests
func TestJoinCommand_NewUser(t *testing.T) {
	userRepo := &mockUserRepository{}
	sessionRepo := &mockSessionRepository{}

	uc := NewJoinCommandUseCase(userRepo, sessionRepo)

	input := JoinCommandInput{
		UserName: "yamada",
		WorkName: "論文執筆",
		Tier:     domain.Tier1,
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.IsNewUser {
		t.Error("expected IsNewUser to be true")
	}
	if output.AlreadyIn {
		t.Error("expected AlreadyIn to be false")
	}
	if output.UserID == 0 {
		t.Error("expected UserID to be set")
	}
	if output.SessionID == 0 {
		t.Error("expected SessionID to be set")
	}
	if output.WorkName != "論文執筆" {
		t.Errorf("expected WorkName to be '論文執筆', got %s", output.WorkName)
	}
}

func TestJoinCommand_ExistingUser(t *testing.T) {
	existingUser := &domain.User{
		ID:        42,
		Name:      "yamada",
		Tier:      domain.Tier2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	userRepo := &mockUserRepository{
		findByNameFn: func(ctx context.Context, name string) (*domain.User, error) {
			if name == "yamada" {
				return existingUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}
	sessionRepo := &mockSessionRepository{}

	uc := NewJoinCommandUseCase(userRepo, sessionRepo)

	input := JoinCommandInput{
		UserName: "yamada",
		WorkName: "コーディング",
		Tier:     domain.Tier1, // This should be ignored since user exists
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.IsNewUser {
		t.Error("expected IsNewUser to be false")
	}
	if output.UserID != 42 {
		t.Errorf("expected UserID to be 42, got %d", output.UserID)
	}
}

func TestJoinCommand_AlreadyActiveSession(t *testing.T) {
	existingUser := &domain.User{
		ID:        42,
		Name:      "yamada",
		Tier:      domain.Tier1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	activeSession := &domain.Session{
		ID:         99,
		UserID:     42,
		WorkName:   "既存の作業",
		StartTime:  time.Now().Add(-30 * time.Minute),
		PlannedEnd: time.Now().Add(30 * time.Minute),
		ActualEnd:  nil,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	userRepo := &mockUserRepository{
		findByNameFn: func(ctx context.Context, name string) (*domain.User, error) {
			return existingUser, nil
		},
	}
	sessionRepo := &mockSessionRepository{
		findActiveByUserIDFn: func(ctx context.Context, userID int64) (*domain.Session, error) {
			if userID == 42 {
				return activeSession, nil
			}
			return nil, errors.New("session not found")
		},
	}

	uc := NewJoinCommandUseCase(userRepo, sessionRepo)

	input := JoinCommandInput{
		UserName: "yamada",
		WorkName: "新しい作業",
		Tier:     domain.Tier1,
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.AlreadyIn {
		t.Error("expected AlreadyIn to be true")
	}
	if output.SessionID != 99 {
		t.Errorf("expected SessionID to be 99, got %d", output.SessionID)
	}
	if output.WorkName != "既存の作業" {
		t.Errorf("expected WorkName to be '既存の作業', got %s", output.WorkName)
	}
}
