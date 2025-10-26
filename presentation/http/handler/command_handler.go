package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yamada-ai/workspace-backend/domain"
	"github.com/yamada-ai/workspace-backend/presentation/http/dto"
	"github.com/yamada-ai/workspace-backend/usecase/command"
)

// CommandHandler handles command-related HTTP requests
type CommandHandler struct {
	joinUseCase *command.JoinCommandUseCase
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(joinUseCase *command.JoinCommandUseCase) *CommandHandler {
	return &CommandHandler{
		joinUseCase: joinUseCase,
	}
}

// JoinCommand handles POST /api/commands/join
func (h *CommandHandler) JoinCommand(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req dto.JoinCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate user_name
	if req.UserName == "" {
		writeError(w, http.StatusBadRequest, "user_name is required")
		return
	}

	// Prepare usecase input
	workName := ""
	if req.WorkName != nil {
		workName = *req.WorkName
	}

	input := command.JoinCommandInput{
		UserName: req.UserName,
		WorkName: workName,
	}

	// Execute usecase
	output, err := h.joinUseCase.Execute(r.Context(), input)
	if err != nil {
		// Handle already in session error
		if err == domain.ErrUserAlreadyInSession {
			writeError(w, http.StatusConflict, "既に作業セッション中です。先に /out で終了してください。")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to join: "+err.Error())
		return
	}

	// Convert to response
	resp := dto.JoinCommandResponse{
		SessionId:  output.SessionID,
		UserId:     output.UserID,
		StartTime:  output.StartTime,
		PlannedEnd: output.PlannedEnd,
	}
	if output.WorkName != "" {
		resp.WorkName = &output.WorkName
	}

	writeJSON(w, http.StatusOK, resp)
}

// HealthCheck handles GET /health
func (h *CommandHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		// Log error but don't fail the health check since headers are already sent
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// Helper functions
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error - headers already sent, so we can't change response
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, dto.ErrorResponse{Error: message})
}
