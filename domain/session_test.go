package domain

import (
	"testing"
	"time"
)

func fixedTime() time.Time {
	return time.Date(2025, 10, 9, 14, 30, 0, 0, time.UTC)
}

func TestNewSession(t *testing.T) {
	userID := int64(1)
	workName := "論文執筆"
	duration := 60 * time.Minute

	session, err := NewSession(userID, workName, duration, fixedTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, session.UserID)
	}
	if session.WorkName != workName {
		t.Errorf("expected WorkName %s, got %s", workName, session.WorkName)
	}
	if !session.StartTime.Equal(fixedTime()) {
		t.Errorf("expected StartTime %v, got %v", fixedTime(), session.StartTime)
	}
	expectedEnd := fixedTime().Add(duration)
	if !session.PlannedEnd.Equal(expectedEnd) {
		t.Errorf("expected PlannedEnd %v, got %v", expectedEnd, session.PlannedEnd)
	}
	if session.ActualEnd != nil {
		t.Errorf("expected ActualEnd to be nil, got %v", session.ActualEnd)
	}
	if !session.IsActive() {
		t.Error("expected session to be active")
	}
}

func TestNewSession_InvalidDuration(t *testing.T) {
	_, err := NewSession(1, "work", 0, fixedTime)
	if err != ErrInvalidDuration {
		t.Errorf("expected ErrInvalidDuration, got %v", err)
	}

	_, err = NewSession(1, "work", -10*time.Minute, fixedTime)
	if err != ErrInvalidDuration {
		t.Errorf("expected ErrInvalidDuration, got %v", err)
	}
}

func TestSession_Extend(t *testing.T) {
	session, _ := NewSession(1, "work", 60*time.Minute, fixedTime)
	originalPlannedEnd := session.PlannedEnd

	err := session.Extend(30*time.Minute, fixedTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedEnd := originalPlannedEnd.Add(30 * time.Minute)
	if !session.PlannedEnd.Equal(expectedEnd) {
		t.Errorf("expected PlannedEnd %v, got %v", expectedEnd, session.PlannedEnd)
	}
}

func TestSession_Extend_InvalidDuration(t *testing.T) {
	session, _ := NewSession(1, "work", 60*time.Minute, fixedTime)

	err := session.Extend(0, fixedTime)
	if err != ErrInvalidExtension {
		t.Errorf("expected ErrInvalidExtension, got %v", err)
	}

	err = session.Extend(-10*time.Minute, fixedTime)
	if err != ErrInvalidExtension {
		t.Errorf("expected ErrInvalidExtension, got %v", err)
	}
}

func TestSession_Extend_CompletedSession(t *testing.T) {
	session, _ := NewSession(1, "work", 60*time.Minute, fixedTime)
	_ = session.Complete(fixedTime)

	err := session.Extend(30*time.Minute, fixedTime)
	if err != ErrSessionAlreadyCompleted {
		t.Errorf("expected ErrSessionAlreadyCompleted, got %v", err)
	}
}

func TestSession_Complete(t *testing.T) {
	session, _ := NewSession(1, "work", 60*time.Minute, fixedTime)

	err := session.Complete(fixedTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.IsActive() {
		t.Error("expected session to be inactive")
	}
	if session.ActualEnd == nil {
		t.Error("expected ActualEnd to be set")
	}
	if !session.ActualEnd.Equal(fixedTime()) {
		t.Errorf("expected ActualEnd %v, got %v", fixedTime(), *session.ActualEnd)
	}
}

func TestSession_Complete_AlreadyCompleted(t *testing.T) {
	session, _ := NewSession(1, "work", 60*time.Minute, fixedTime)
	_ = session.Complete(fixedTime)

	err := session.Complete(fixedTime)
	if err != ErrSessionAlreadyCompleted {
		t.Errorf("expected ErrSessionAlreadyCompleted, got %v", err)
	}
}

func TestSession_Duration(t *testing.T) {
	startTime := fixedTime()
	session, _ := NewSession(1, "work", 60*time.Minute, func() time.Time { return startTime })

	// Active session: duration from start to now
	now := func() time.Time { return startTime.Add(30 * time.Minute) }
	duration := session.Duration(now)
	if duration != 30*time.Minute {
		t.Errorf("expected duration 30m, got %v", duration)
	}

	// Completed session: duration from start to actual_end
	_ = session.Complete(now)
	duration = session.Duration(now)
	if duration != 30*time.Minute {
		t.Errorf("expected duration 30m, got %v", duration)
	}

	// Even if "now" moves forward, duration should remain the same for completed session
	laterNow := func() time.Time { return startTime.Add(90 * time.Minute) }
	duration = session.Duration(laterNow)
	if duration != 30*time.Minute {
		t.Errorf("expected duration 30m, got %v", duration)
	}
}
