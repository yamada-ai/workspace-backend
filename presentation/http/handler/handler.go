package handler

import "net/http"

// Handler combines all HTTP handlers (commands and queries)
type Handler struct {
	*CommandHandler
	*QueryHandler
}

// NewHandler creates a unified handler that implements dto.ServerInterface
func NewHandler(
	commandHandler *CommandHandler,
	queryHandler *QueryHandler,
) *Handler {
	return &Handler{
		CommandHandler: commandHandler,
		QueryHandler:   queryHandler,
	}
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
