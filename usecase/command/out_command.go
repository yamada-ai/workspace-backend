package command

import (
	"context"
	"time"

	"github.com/yamada-ai/workspace-backend/domain"
	"github.com/yamada-ai/workspace-backend/domain/repository"
)

// CompleteSessionService defines the interface for completing sessions
type CompleteSessionService interface {
	CompleteSession(ctx context.Context, session *domain.Session, userID int64) error
}

// ExpirationCanceller defines the interface for cancelling session expiration
type ExpirationCanceller interface {
	CancelExpiration(sessionID int64)
}

// OutCommandInput represents the input for out command
type OutCommandInput struct {
	UserName string
}

// OutCommandOutput represents the output of out command
type OutCommandOutput struct {
	SessionID int64
	UserID    int64
	ActualEnd time.Time
}

// OutCommandUseCase handles the /out command logic
type OutCommandUseCase struct {
	userRepository      repository.UserRepository
	sessionRepository   repository.SessionRepository
	completeService     CompleteSessionService
	expirationCanceller ExpirationCanceller
}

// NewOutCommandUseCase creates a new out command use case
func NewOutCommandUseCase(
	userRepository repository.UserRepository,
	sessionRepository repository.SessionRepository,
	completeService CompleteSessionService,
	expirationCanceller ExpirationCanceller,
) *OutCommandUseCase {
	return &OutCommandUseCase{
		userRepository:      userRepository,
		sessionRepository:   sessionRepository,
		completeService:     completeService,
		expirationCanceller: expirationCanceller,
	}
}

// Execute executes the out command
func (uc *OutCommandUseCase) Execute(ctx context.Context, input OutCommandInput) (*OutCommandOutput, error) {
	// 1. Find user
	user, err := uc.userRepository.FindByName(ctx, input.UserName)
	if err != nil {
		return nil, err
	}

	// 2. Find active session
	session, err := uc.sessionRepository.FindActiveByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// 3. Complete the session using shared service (Complete + Update + Broadcast)
	if err := uc.completeService.CompleteSession(ctx, session, user.ID); err != nil {
		return nil, err
	}

	// 4. Cancel the automatic expiration timer
	uc.expirationCanceller.CancelExpiration(session.ID)

	return &OutCommandOutput{
		SessionID: session.ID,
		UserID:    user.ID,
		ActualEnd: *session.ActualEnd,
	}, nil
}
