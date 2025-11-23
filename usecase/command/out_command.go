package command

import (
	"context"
	"time"

	"github.com/yamada-ai/workspace-backend/domain/repository"
)

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
	userRepository    repository.UserRepository
	sessionRepository repository.SessionRepository
	broadcaster       EventBroadcaster
	now               func() time.Time
}

// NewOutCommandUseCase creates a new out command use case
func NewOutCommandUseCase(
	userRepository repository.UserRepository,
	sessionRepository repository.SessionRepository,
	broadcaster EventBroadcaster,
) *OutCommandUseCase {
	return &OutCommandUseCase{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		broadcaster:       broadcaster,
		now:               time.Now,
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

	// 3. Complete the session
	if err := session.Complete(uc.now); err != nil {
		return nil, err
	}

	// 4. Update session in database
	if err := uc.sessionRepository.Update(ctx, session); err != nil {
		return nil, err
	}

	// 5. Broadcast session end event to all connected WebSocket clients
	uc.broadcaster.BroadcastSessionEnd(SessionEndBroadcast{
		SessionID: session.ID,
		UserID:    user.ID,
		ActualEnd: *session.ActualEnd,
	})

	return &OutCommandOutput{
		SessionID: session.ID,
		UserID:    user.ID,
		ActualEnd: *session.ActualEnd,
	}, nil
}
