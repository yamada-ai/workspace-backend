package command

import (
	"context"
	"time"

	"github.com/yamada-ai/workspace-backend/domain"
	"github.com/yamada-ai/workspace-backend/domain/repository"
)

const (
	// MinExtensionMinutes 最小延長時間（分）
	MinExtensionMinutes = 1
	// MaxExtensionMinutes 最大延長時間（分）
	MaxExtensionMinutes = 360
)

// ExpirationRescheduler defines the interface for rescheduling session expiration
type ExpirationRescheduler interface {
	RescheduleExpiration(sessionID int64, userID int64, newPlannedEnd time.Time)
}

// MoreCommandInput represents the input for more command
type MoreCommandInput struct {
	UserName string
	Minutes  int
}

// MoreCommandOutput represents the output of more command
type MoreCommandOutput struct {
	SessionID  int64
	UserID     int64
	Minutes    int
	PlannedEnd time.Time
}

// MoreCommandUseCase handles the /more command logic
type MoreCommandUseCase struct {
	userRepository      repository.UserRepository
	sessionRepository   repository.SessionRepository
	expirationScheduler ExpirationRescheduler
	now                 func() time.Time
}

// NewMoreCommandUseCase creates a new more command use case
func NewMoreCommandUseCase(
	userRepository repository.UserRepository,
	sessionRepository repository.SessionRepository,
	expirationScheduler ExpirationRescheduler,
) *MoreCommandUseCase {
	return &MoreCommandUseCase{
		userRepository:      userRepository,
		sessionRepository:   sessionRepository,
		expirationScheduler: expirationScheduler,
		now:                 func() time.Time { return time.Now().UTC() },
	}
}

// Execute executes the more command
func (uc *MoreCommandUseCase) Execute(ctx context.Context, input MoreCommandInput) (*MoreCommandOutput, error) {
	// 1. Validate minutes
	if input.Minutes < MinExtensionMinutes || input.Minutes > MaxExtensionMinutes {
		return nil, domain.ErrInvalidExtension
	}

	// 2. Find user
	user, err := uc.userRepository.FindByName(ctx, input.UserName)
	if err != nil {
		return nil, err
	}

	// 3. Find active session
	session, err := uc.sessionRepository.FindActiveByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// 4. Extend the session
	duration := time.Duration(input.Minutes) * time.Minute
	if err := session.Extend(duration, uc.now); err != nil {
		return nil, err
	}

	// 5. Update session in database
	if err := uc.sessionRepository.Update(ctx, session); err != nil {
		return nil, err
	}

	// 6. Reschedule expiration timer
	uc.expirationScheduler.RescheduleExpiration(session.ID, user.ID, session.PlannedEnd)

	return &MoreCommandOutput{
		SessionID:  session.ID,
		UserID:     user.ID,
		Minutes:    input.Minutes,
		PlannedEnd: session.PlannedEnd,
	}, nil
}
