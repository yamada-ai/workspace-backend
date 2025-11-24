package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yamada-ai/workspace-backend/domain"
	"github.com/yamada-ai/workspace-backend/domain/repository"
)

// Mock Transaction
type mockTx struct {
	commitFn   func(ctx context.Context) error
	rollbackFn func(ctx context.Context) error
}

func (m *mockTx) Commit(ctx context.Context) error {
	if m.commitFn != nil {
		return m.commitFn(ctx)
	}
	return nil
}

func (m *mockTx) Rollback(ctx context.Context) error {
	if m.rollbackFn != nil {
		return m.rollbackFn(ctx)
	}
	return nil
}

// Mock UserRepository
type mockUserRepository struct {
	findByNameFn       func(ctx context.Context, name string) (*domain.User, error)
	findByIDFn         func(ctx context.Context, id int64) (*domain.User, error)
	saveFn             func(ctx context.Context, user *domain.User) error
	beginTxFn          func(ctx context.Context) (repository.Tx, error)
	findByNameWithTxFn func(ctx context.Context, tx repository.Tx, name string) (*domain.User, error)
	saveWithTxFn       func(ctx context.Context, tx repository.Tx, user *domain.User) error
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

func (m *mockUserRepository) BeginTx(ctx context.Context) (repository.Tx, error) {
	if m.beginTxFn != nil {
		return m.beginTxFn(ctx)
	}
	return &mockTx{}, nil
}

func (m *mockUserRepository) FindByNameWithTx(ctx context.Context, tx repository.Tx, name string) (*domain.User, error) {
	if m.findByNameWithTxFn != nil {
		return m.findByNameWithTxFn(ctx, tx, name)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) SaveWithTx(ctx context.Context, tx repository.Tx, user *domain.User) error {
	if m.saveWithTxFn != nil {
		return m.saveWithTxFn(ctx, tx, user)
	}
	// Simulate ID assignment
	if user.ID == 0 {
		user.ID = 1
	}
	return nil
}

// Mock SessionRepository
type mockSessionRepository struct {
	saveFn                     func(ctx context.Context, session *domain.Session) error
	createFn                   func(ctx context.Context, session *domain.Session) error
	findByIDFn                 func(ctx context.Context, id int64) (*domain.Session, error)
	findActiveByUserIDFn       func(ctx context.Context, userID int64) (*domain.Session, error)
	updateFn                   func(ctx context.Context, session *domain.Session) error
	listByUserIDFn             func(ctx context.Context, userID int64, limit, offset int32) ([]*domain.Session, error)
	findActiveByUserIDWithTxFn func(ctx context.Context, tx repository.Tx, userID int64) (*domain.Session, error)
	createWithTxFn             func(ctx context.Context, tx repository.Tx, session *domain.Session) error
	findAllActiveFn            func(ctx context.Context) ([]domain.SessionInfo, error)
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

func (m *mockSessionRepository) FindActiveByUserIDWithTx(ctx context.Context, tx repository.Tx, userID int64) (*domain.Session, error) {
	if m.findActiveByUserIDWithTxFn != nil {
		return m.findActiveByUserIDWithTxFn(ctx, tx, userID)
	}
	return nil, domain.ErrSessionNotFound
}

func (m *mockSessionRepository) CreateWithTx(ctx context.Context, tx repository.Tx, session *domain.Session) error {
	if m.createWithTxFn != nil {
		return m.createWithTxFn(ctx, tx, session)
	}
	// Simulate ID assignment
	if session.ID == 0 {
		session.ID = 100
	}
	return nil
}

func (m *mockSessionRepository) FindAllActive(ctx context.Context) ([]domain.SessionInfo, error) {
	if m.findAllActiveFn != nil {
		return m.findAllActiveFn(ctx)
	}
	return []domain.SessionInfo{}, nil
}

func (m *mockSessionRepository) FindByUserIDAndDateRange(ctx context.Context, userID int64, startTime, endTime time.Time) ([]*domain.Session, error) {
	return nil, nil
}

// Tests
func TestJoinCommand_NewUser(t *testing.T) {
	userRepository := &mockUserRepository{}
	sessionRepository := &mockSessionRepository{}

	uc := NewJoinCommandUseCase(userRepository, sessionRepository, NoOpBroadcaster{}, NoOpExpirationScheduler{})

	input := JoinCommandInput{
		UserName: "yamada",
		WorkName: "論文執筆",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.IsNewUser {
		t.Error("expected IsNewUser to be true")
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
		findByNameWithTxFn: func(ctx context.Context, tx repository.Tx, name string) (*domain.User, error) {
			if name == "yamada" {
				return existingUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}
	sessionRepo := &mockSessionRepository{}

	uc := NewJoinCommandUseCase(userRepo, sessionRepo, NoOpBroadcaster{}, NoOpExpirationScheduler{})

	input := JoinCommandInput{
		UserName: "yamada",
		WorkName: "コーディング",
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
		findByNameWithTxFn: func(ctx context.Context, tx repository.Tx, name string) (*domain.User, error) {
			return existingUser, nil
		},
	}
	sessionRepo := &mockSessionRepository{
		findActiveByUserIDWithTxFn: func(ctx context.Context, tx repository.Tx, userID int64) (*domain.Session, error) {
			if userID == 42 {
				return activeSession, nil
			}
			return nil, domain.ErrSessionNotFound
		},
	}

	uc := NewJoinCommandUseCase(userRepo, sessionRepo, NoOpBroadcaster{}, NoOpExpirationScheduler{})

	input := JoinCommandInput{
		UserName: "yamada",
		WorkName: "新しい作業",
	}

	output, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	if !errors.Is(err, domain.ErrUserAlreadyInSession) {
		t.Errorf("expected ErrUserAlreadyInSession, got %v", err)
	}

	if output != nil {
		t.Errorf("expected output to be nil when error occurs, got %+v", output)
	}
}
