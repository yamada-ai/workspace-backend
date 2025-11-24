package command

import (
	"context"
	"time"

	"github.com/yamada-ai/workspace-backend/domain/repository"
)

// WorkNameChangeBroadcaster defines the interface for broadcasting work name changes
type WorkNameChangeBroadcaster interface {
	BroadcastWorkNameChange(event WorkNameChangeBroadcast)
}

// ChangeCommandInput represents the input for change command
type ChangeCommandInput struct {
	UserName    string
	NewWorkName string
}

// ChangeCommandOutput represents the output of change command
type ChangeCommandOutput struct {
	SessionID int64
	UserID    int64
	WorkName  string
}

// ChangeCommandUseCase handles the /change command logic
type ChangeCommandUseCase struct {
	userRepository    repository.UserRepository
	sessionRepository repository.SessionRepository
	broadcaster       WorkNameChangeBroadcaster
	now               func() time.Time
}

// NewChangeCommandUseCase creates a new change command use case
func NewChangeCommandUseCase(
	userRepository repository.UserRepository,
	sessionRepository repository.SessionRepository,
	broadcaster WorkNameChangeBroadcaster,
) *ChangeCommandUseCase {
	return &ChangeCommandUseCase{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		broadcaster:       broadcaster,
		now:               func() time.Time { return time.Now().UTC() },
	}
}

// Execute executes the change command
func (uc *ChangeCommandUseCase) Execute(ctx context.Context, input ChangeCommandInput) (*ChangeCommandOutput, error) {
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

	// 3. Change work name
	if err := session.ChangeWorkName(input.NewWorkName, uc.now); err != nil {
		return nil, err
	}

	// 4. Update session in database
	if err := uc.sessionRepository.Update(ctx, session); err != nil {
		return nil, err
	}

	// 5. Broadcast work name change event
	uc.broadcaster.BroadcastWorkNameChange(WorkNameChangeBroadcast{
		SessionID: session.ID,
		UserID:    user.ID,
		WorkName:  session.WorkName,
	})

	return &ChangeCommandOutput{
		SessionID: session.ID,
		UserID:    user.ID,
		WorkName:  session.WorkName,
	}, nil
}
