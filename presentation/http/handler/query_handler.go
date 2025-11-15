package handler

import (
	"net/http"

	"github.com/yamada-ai/workspace-backend/presentation/http/dto"
	"github.com/yamada-ai/workspace-backend/usecase/query"
)

// QueryHandler handles query (read) requests
type QueryHandler struct {
	getActiveSessionsUseCase *query.GetActiveSessionsUseCase
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(
	getActiveSessionsUseCase *query.GetActiveSessionsUseCase,
) *QueryHandler {
	return &QueryHandler{
		getActiveSessionsUseCase: getActiveSessionsUseCase,
	}
}

// GetActiveSessions handles GET /api/sessions/active
// (GET /api/sessions/active)
func (h *QueryHandler) GetActiveSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Execute use case
	output, err := h.getActiveSessionsUseCase.Execute(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to retrieve active sessions")
		return
	}

	// Convert to DTO
	sessions := make([]dto.SessionInfo, 0, len(output.Sessions))
	for _, s := range output.Sessions {
		sessions = append(sessions, dto.SessionInfo{
			SessionId:  s.SessionID,
			UserId:     s.UserID,
			UserName:   s.UserName,
			WorkName:   s.WorkName,
			Tier:       s.Tier,
			IconId:     s.IconID,
			StartTime:  s.StartTime,
			PlannedEnd: s.PlannedEnd,
		})
	}

	response := dto.ActiveSessionsResponse{
		Sessions: sessions,
	}

	writeJSON(w, http.StatusOK, response)
}
