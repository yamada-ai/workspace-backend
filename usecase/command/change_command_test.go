package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yamada-ai/workspace-backend/domain"
)

func TestChangeCommand_Success(t *testing.T) {
	now := time.Now()

	existingUser := &domain.User{
		ID:        42,
		Name:      "yamada",
		Tier:      domain.Tier1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	activeSession := &domain.Session{
		ID:         99,
		UserID:     42,
		WorkName:   "論文執筆",
		StartTime:  now.Add(-30 * time.Minute),
		PlannedEnd: now.Add(30 * time.Minute),
		ActualEnd:  nil,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	userRepo := &mockUserRepository{
		findByNameFn: func(ctx context.Context, name string) (*domain.User, error) {
			if name == "yamada" {
				return existingUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}

	sessionUpdated := false
	sessionRepo := &mockSessionRepository{
		findActiveByUserIDFn: func(ctx context.Context, userID int64) (*domain.Session, error) {
			if userID == 42 {
				return activeSession, nil
			}
			return nil, domain.ErrSessionNotFound
		},
		updateFn: func(ctx context.Context, session *domain.Session) error {
			sessionUpdated = true
			return nil
		},
	}

	broadcastCalled := false
	broadcaster := &mockWorkNameChangeBroadcaster{
		broadcastWorkNameChangeFn: func(event WorkNameChangeBroadcast) {
			if event.SessionID != 99 || event.UserID != 42 || event.WorkName != "資格勉強" {
				t.Errorf("unexpected broadcast: %+v", event)
			}
			broadcastCalled = true
		},
	}

	uc := NewChangeCommandUseCase(userRepo, sessionRepo, broadcaster)

	input := ChangeCommandInput{
		UserName:    "yamada",
		NewWorkName: "資格勉強",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.SessionID != 99 {
		t.Errorf("expected SessionID to be 99, got %d", output.SessionID)
	}
	if output.UserID != 42 {
		t.Errorf("expected UserID to be 42, got %d", output.UserID)
	}
	if output.WorkName != "資格勉強" {
		t.Errorf("expected WorkName to be '資格勉強', got %s", output.WorkName)
	}

	if !sessionUpdated {
		t.Error("expected session to be updated")
	}

	if !broadcastCalled {
		t.Error("expected work name change to be broadcast")
	}
}

func TestChangeCommand_EmptyWorkName(t *testing.T) {
	now := time.Now()

	existingUser := &domain.User{
		ID:        42,
		Name:      "yamada",
		Tier:      domain.Tier1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	activeSession := &domain.Session{
		ID:         99,
		UserID:     42,
		WorkName:   "論文執筆",
		StartTime:  now.Add(-30 * time.Minute),
		PlannedEnd: now.Add(30 * time.Minute),
		ActualEnd:  nil,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	userRepo := &mockUserRepository{
		findByNameFn: func(ctx context.Context, name string) (*domain.User, error) {
			return existingUser, nil
		},
	}

	sessionRepo := &mockSessionRepository{
		findActiveByUserIDFn: func(ctx context.Context, userID int64) (*domain.Session, error) {
			return activeSession, nil
		},
		updateFn: func(ctx context.Context, session *domain.Session) error {
			return nil
		},
	}

	broadcaster := &mockWorkNameChangeBroadcaster{
		broadcastWorkNameChangeFn: func(event WorkNameChangeBroadcast) {},
	}

	uc := NewChangeCommandUseCase(userRepo, sessionRepo, broadcaster)

	input := ChangeCommandInput{
		UserName:    "yamada",
		NewWorkName: "", // Empty work name should be allowed
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.WorkName != "" {
		t.Errorf("expected WorkName to be empty, got %s", output.WorkName)
	}
}

func TestChangeCommand_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepository{
		findByNameFn: func(ctx context.Context, name string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	sessionRepo := &mockSessionRepository{}
	broadcaster := &mockWorkNameChangeBroadcaster{}

	uc := NewChangeCommandUseCase(userRepo, sessionRepo, broadcaster)

	input := ChangeCommandInput{
		UserName:    "nonexistent",
		NewWorkName: "作業",
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

func TestChangeCommand_NoActiveSession(t *testing.T) {
	existingUser := &domain.User{
		ID:        42,
		Name:      "yamada",
		Tier:      domain.Tier1,
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

	sessionRepo := &mockSessionRepository{
		findActiveByUserIDFn: func(ctx context.Context, userID int64) (*domain.Session, error) {
			return nil, domain.ErrSessionNotFound
		},
	}

	broadcaster := &mockWorkNameChangeBroadcaster{}

	uc := NewChangeCommandUseCase(userRepo, sessionRepo, broadcaster)

	input := ChangeCommandInput{
		UserName:    "yamada",
		NewWorkName: "作業",
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

// Mock WorkNameChangeBroadcaster
type mockWorkNameChangeBroadcaster struct {
	broadcastWorkNameChangeFn func(event WorkNameChangeBroadcast)
}

func (m *mockWorkNameChangeBroadcaster) BroadcastWorkNameChange(event WorkNameChangeBroadcast) {
	if m.broadcastWorkNameChangeFn != nil {
		m.broadcastWorkNameChangeFn(event)
	}
}
