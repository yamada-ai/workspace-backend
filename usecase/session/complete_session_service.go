package session

import (
	"context"
	"time"

	"github.com/yamada-ai/workspace-backend/domain"
	"github.com/yamada-ai/workspace-backend/domain/repository"
	"github.com/yamada-ai/workspace-backend/usecase/command"
)

// CompleteSessionService handles the common logic for completing a session
type CompleteSessionService struct {
	sessionRepository repository.SessionRepository
	broadcaster       command.EventBroadcaster
	now               func() time.Time
}

// NewCompleteSessionService creates a new complete session service
func NewCompleteSessionService(
	sessionRepository repository.SessionRepository,
	broadcaster command.EventBroadcaster,
) *CompleteSessionService {
	return &CompleteSessionService{
		sessionRepository: sessionRepository,
		broadcaster:       broadcaster,
		now:               func() time.Time { return time.Now().UTC() },
	}
}

// CompleteSession completes a session, updates it in the database, and broadcasts the event
// This method is used by both manual /out command and automatic expiration
func (s *CompleteSessionService) CompleteSession(
	ctx context.Context,
	session *domain.Session,
	userID int64,
) error {
	// 1. Complete the session (sets actual_end)
	if err := session.Complete(s.now); err != nil {
		return err
	}

	// 2. Update session in database
	if err := s.sessionRepository.Update(ctx, session); err != nil {
		return err
	}

	// 3. Broadcast session end event to all connected WebSocket clients
	s.broadcaster.BroadcastSessionEnd(command.SessionEndBroadcast{
		SessionID: session.ID,
		UserID:    userID,
		ActualEnd: *session.ActualEnd,
	})

	return nil
}
