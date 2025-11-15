package query

import (
	"context"

	"github.com/yamada-ai/workspace-backend/domain"
	"github.com/yamada-ai/workspace-backend/domain/repository"
)

// GetActiveSessionsOutput represents the output of GetActiveSessions query
type GetActiveSessionsOutput struct {
	Sessions []domain.SessionInfo `json:"sessions"`
}

// GetActiveSessionsUseCase handles retrieving all active sessions
type GetActiveSessionsUseCase struct {
	sessionRepository repository.SessionRepository
}

// NewGetActiveSessionsUseCase creates a new use case instance
func NewGetActiveSessionsUseCase(
	sessionRepository repository.SessionRepository,
) *GetActiveSessionsUseCase {
	return &GetActiveSessionsUseCase{
		sessionRepository: sessionRepository,
	}
}

// Execute retrieves all active sessions with user information
func (uc *GetActiveSessionsUseCase) Execute(ctx context.Context) (*GetActiveSessionsOutput, error) {
	sessions, err := uc.sessionRepository.FindAllActive(ctx)
	if err != nil {
		return nil, err
	}

	return &GetActiveSessionsOutput{
		Sessions: sessions,
	}, nil
}
