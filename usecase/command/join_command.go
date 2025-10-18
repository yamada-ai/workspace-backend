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
	userRepository    repository.UserRepository
	sessionRepository repository.SessionRepository
	now               func() time.Time
}

// NewJoinCommandUseCase creates a new join command use case
func NewJoinCommandUseCase(
	userRepository repository.UserRepository,
	sessionRepository repository.SessionRepository,
) *JoinCommandUseCase {
	return &JoinCommandUseCase{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		now:               time.Now,
	}
}

// Execute executes the join command
func (uc *JoinCommandUseCase) Execute(ctx context.Context, input JoinCommandInput) (*JoinCommandOutput, error) {
	// Start transaction
	tx, err := uc.userRepository.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// 1. Find or create user within transaction
	user, err := uc.userRepository.FindByNameWithTx(ctx, tx, input.UserName)
	isNewUser := false
	if err == domain.ErrUserNotFound {
		// Create new user with default Tier1
		user, err = domain.NewUser(input.UserName, domain.Tier1, uc.now)
		if err != nil {
			return nil, err
		}
		if err := uc.userRepository.SaveWithTx(ctx, tx, user); err != nil {
			if err == domain.ErrUserAlreadyExists {
				// Another goroutine created the user, retry find
				user, err = uc.userRepository.FindByNameWithTx(ctx, tx, input.UserName)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		} else {
			isNewUser = true
		}
	} else if err != nil {
		return nil, err
	}

	// 2. Check if user already has an active session
	activeSession, err := uc.sessionRepository.FindActiveByUserIDWithTx(ctx, tx, user.ID)
	if err == nil {
		// User already has an active session, commit and return
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return &JoinCommandOutput{
			SessionID:  activeSession.ID,
			UserID:     user.ID,
			WorkName:   activeSession.WorkName,
			StartTime:  activeSession.StartTime,
			PlannedEnd: activeSession.PlannedEnd,
			IsNewUser:  isNewUser,
			AlreadyIn:  true,
		}, nil
	} else if err != domain.ErrSessionNotFound {
		return nil, err
	}

	// 3. Create new session
	session, err := domain.NewSession(user.ID, input.WorkName, DefaultSessionDuration, uc.now)
	if err != nil {
		return nil, err
	}

	if err := uc.sessionRepository.CreateWithTx(ctx, tx, session); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
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
