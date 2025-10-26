package domain

import (
	"errors"
	"time"
)

var (
	ErrSessionNotFound         = errors.New("session not found")
	ErrSessionAlreadyCompleted = errors.New("session already completed")
	ErrInvalidDuration         = errors.New("invalid duration: must be positive")
	ErrInvalidExtension        = errors.New("invalid extension: must be positive")
	ErrUserAlreadyInSession    = errors.New("user already has an active session")
)

// Session 作業セッションを表す
type Session struct {
	ID         int64
	UserID     int64
	WorkName   string
	StartTime  time.Time
	PlannedEnd time.Time
	ActualEnd  *time.Time // セッションがアクティブな場合はnil
	IconID     *int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewSession 新しい作業セッションを作成する
// duration はデフォルトの作業時間（例: 60分）
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
		ActualEnd:  nil, // アクティブなセッション
		CreatedAt:  nowT,
		UpdatedAt:  nowT,
	}, nil
}

// IsActive セッションがまだアクティブか（完了していないか）を確認する
func (s *Session) IsActive() bool {
	return s.ActualEnd == nil
}

// Extend 予定終了時刻を指定した時間だけ延長する
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

// Complete セッションを完了としてマークする
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

// Duration セッションの実際の継続時間を返す
// セッションがまだアクティブな場合は、開始から現在までの時間を返す
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

// Touch updated_atのタイムスタンプを更新する
func (s *Session) Touch(now func() time.Time) {
	t := time.Now
	if now != nil {
		t = now
	}
	s.UpdatedAt = t()
}
