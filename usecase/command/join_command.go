package command

import (
	"context"
	"time"

	"github.com/yamada-ai/workspace-backend/domain"
	"github.com/yamada-ai/workspace-backend/domain/repository"
)

const (
	// DefaultSessionDuration is the default work session duration (60 minutes)
	DefaultSessionDuration = 60 * time.Minute
)

// JoinCommandInput represents the input for join command
type JoinCommandInput struct {
	UserName string
	WorkName string
	Tier     domain.Tier
}

// JoinCommandOutput represents the output of join command
type JoinCommandOutput struct {
	SessionID  int64
	UserID     int64
	WorkName   string
	StartTime  time.Time
	PlannedEnd time.Time
	IsNewUser  bool
	AlreadyIn  bool // true if user already has an active session
}

// JoinCommandUseCase handles the /in command logic
type JoinCommandUseCase struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	now         func() time.Time
}

// NewJoinCommandUseCase creates a new join command use case
func NewJoinCommandUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
) *JoinCommandUseCase {
	return &JoinCommandUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		now:         time.Now,
	}
}

// Execute executes the join command
func (uc *JoinCommandUseCase) Execute(ctx context.Context, input JoinCommandInput) (*JoinCommandOutput, error) {
	// 1. Find or create user
	user, err := uc.userRepo.FindByName(ctx, input.UserName)
	isNewUser := false
	if err == domain.ErrUserNotFound {
		// Create new user
		user, err = domain.NewUser(input.UserName, input.Tier, uc.now)
		if err != nil {
			return nil, err
		}
		if err := uc.userRepo.Save(ctx, user); err != nil {
			return nil, err
		}
		isNewUser = true
	} else if err != nil {
		return nil, err
	}

	// 2. Check if user already has an active session
	activeSession, err := uc.sessionRepo.FindActiveByUserID(ctx, user.ID)
	if err == nil {
		// User already has an active session
		return &JoinCommandOutput{
			SessionID:  activeSession.ID,
			UserID:     user.ID,
			WorkName:   activeSession.WorkName,
			StartTime:  activeSession.StartTime,
			PlannedEnd: activeSession.PlannedEnd,
			IsNewUser:  isNewUser,
			AlreadyIn:  true,
		}, nil
	}

	// 3. Create new session
	session, err := domain.NewSession(user.ID, input.WorkName, DefaultSessionDuration, uc.now)
	if err != nil {
		return nil, err
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &JoinCommandOutput{
		SessionID:  session.ID,
		UserID:     user.ID,
		WorkName:   session.WorkName,
		StartTime:  session.StartTime,
		PlannedEnd: session.PlannedEnd,
		IsNewUser:  isNewUser,
		AlreadyIn:  false,
	}, nil
}
