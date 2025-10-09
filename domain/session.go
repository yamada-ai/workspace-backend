package domain

import (
	"errors"
	"time"
)

var (
	ErrSessionAlreadyCompleted = errors.New("session already completed")
	ErrInvalidDuration         = errors.New("invalid duration: must be positive")
	ErrInvalidExtension        = errors.New("invalid extension: must be positive")
)

// Session represents a work session
type Session struct {
	ID          int64
	UserID      int64
	WorkName    string
	StartTime   time.Time
	PlannedEnd  time.Time
	ActualEnd   *time.Time // nil if session is still active
	IconID      *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewSession creates a new work session
// duration is the default work duration (e.g., 60 minutes)
func NewSession(userID int64, workName string, duration time.Duration, now func() time.Time) (*Session, error) {
	if duration <= 0 {
		return nil, ErrInvalidDuration
	}

	t := time.Now
	if now != nil {
		t = now
	}
	nowT := t()

	return &Session{
		UserID:     userID,
		WorkName:   workName,
		StartTime:  nowT,
		PlannedEnd: nowT.Add(duration),
		ActualEnd:  nil, // active session
		CreatedAt:  nowT,
		UpdatedAt:  nowT,
	}, nil
}

// IsActive checks if the session is still active (not completed)
func (s *Session) IsActive() bool {
	return s.ActualEnd == nil
}

// Extend extends the planned end time by the given duration
func (s *Session) Extend(duration time.Duration, now func() time.Time) error {
	if !s.IsActive() {
		return ErrSessionAlreadyCompleted
	}
	if duration <= 0 {
		return ErrInvalidExtension
	}

	s.PlannedEnd = s.PlannedEnd.Add(duration)
	s.Touch(now)
	return nil
}

// Complete marks the session as completed
func (s *Session) Complete(now func() time.Time) error {
	if !s.IsActive() {
		return ErrSessionAlreadyCompleted
	}

	t := time.Now
	if now != nil {
		t = now
	}
	nowT := t()

	s.ActualEnd = &nowT
	s.UpdatedAt = nowT
	return nil
}

// Duration returns the actual duration of the session
// If the session is still active, returns the duration from start to now
func (s *Session) Duration(now func() time.Time) time.Duration {
	t := time.Now
	if now != nil {
		t = now
	}

	if s.ActualEnd != nil {
		return s.ActualEnd.Sub(s.StartTime)
	}
	return t().Sub(s.StartTime)
}

// Touch updates the updated_at timestamp
func (s *Session) Touch(now func() time.Time) {
	t := time.Now
	if now != nil {
		t = now
	}
	s.UpdatedAt = t()
}
