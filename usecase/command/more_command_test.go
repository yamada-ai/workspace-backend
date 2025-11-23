package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yamada-ai/workspace-backend/domain"
)

func TestMoreCommand_Success(t *testing.T) {
	now := time.Now()
	plannedEnd := now.Add(60 * time.Minute)

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
		PlannedEnd: plannedEnd,
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

	rescheduleCalled := false
	expirationScheduler := &mockExpirationRescheduler{
		rescheduleExpirationFn: func(sessionID int64, userID int64, newPlannedEnd time.Time) {
			if sessionID != 99 || userID != 42 {
				t.Errorf("unexpected reschedule call: sessionID=%d, userID=%d", sessionID, userID)
			}
			rescheduleCalled = true
		},
	}

	uc := NewMoreCommandUseCase(userRepo, sessionRepo, expirationScheduler)

	input := MoreCommandInput{
		UserName: "yamada",
		Minutes:  30,
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
	if output.Minutes != 30 {
		t.Errorf("expected Minutes to be 30, got %d", output.Minutes)
	}

	// Check that PlannedEnd was extended by 30 minutes
	expectedPlannedEnd := plannedEnd.Add(30 * time.Minute)
	if !output.PlannedEnd.Equal(expectedPlannedEnd) {
		t.Errorf("expected PlannedEnd to be %v, got %v", expectedPlannedEnd, output.PlannedEnd)
	}

	if !sessionUpdated {
		t.Error("expected session to be updated")
	}

	if !rescheduleCalled {
		t.Error("expected expiration to be rescheduled")
	}
}

func TestMoreCommand_InvalidMinutes_TooSmall(t *testing.T) {
	userRepo := &mockUserRepository{}
	sessionRepo := &mockSessionRepository{}
	expirationScheduler := &mockExpirationRescheduler{}

	uc := NewMoreCommandUseCase(userRepo, sessionRepo, expirationScheduler)

	input := MoreCommandInput{
		UserName: "yamada",
		Minutes:  0, // Too small
	}

	output, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	if !errors.Is(err, domain.ErrInvalidExtension) {
		t.Errorf("expected ErrInvalidExtension, got %v", err)
	}

	if output != nil {
		t.Errorf("expected output to be nil when error occurs, got %+v", output)
	}
}

func TestMoreCommand_InvalidMinutes_TooLarge(t *testing.T) {
	userRepo := &mockUserRepository{}
	sessionRepo := &mockSessionRepository{}
	expirationScheduler := &mockExpirationRescheduler{}

	uc := NewMoreCommandUseCase(userRepo, sessionRepo, expirationScheduler)

	input := MoreCommandInput{
		UserName: "yamada",
		Minutes:  361, // Too large
	}

	output, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	if !errors.Is(err, domain.ErrInvalidExtension) {
		t.Errorf("expected ErrInvalidExtension, got %v", err)
	}

	if output != nil {
		t.Errorf("expected output to be nil when error occurs, got %+v", output)
	}
}

func TestMoreCommand_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepository{
		findByNameFn: func(ctx context.Context, name string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	sessionRepo := &mockSessionRepository{}
	expirationScheduler := &mockExpirationRescheduler{}

	uc := NewMoreCommandUseCase(userRepo, sessionRepo, expirationScheduler)

	input := MoreCommandInput{
		UserName: "nonexistent",
		Minutes:  30,
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

func TestMoreCommand_NoActiveSession(t *testing.T) {
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

	expirationScheduler := &mockExpirationRescheduler{}

	uc := NewMoreCommandUseCase(userRepo, sessionRepo, expirationScheduler)

	input := MoreCommandInput{
		UserName: "yamada",
		Minutes:  30,
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

// Mock ExpirationRescheduler
type mockExpirationRescheduler struct {
	rescheduleExpirationFn func(sessionID int64, userID int64, newPlannedEnd time.Time)
}

func (m *mockExpirationRescheduler) RescheduleExpiration(sessionID int64, userID int64, newPlannedEnd time.Time) {
	if m.rescheduleExpirationFn != nil {
		m.rescheduleExpirationFn(sessionID, userID, newPlannedEnd)
	}
}
