package domain

import "time"

// SessionInfo represents core session information
// Used across HTTP responses, WebSocket events, and queries
type SessionInfo struct {
	SessionID  int64     `json:"session_id"`
	UserID     int64     `json:"user_id"`
	UserName   string    `json:"user_name"`
	WorkName   string    `json:"work_name"`
	Tier       int       `json:"tier"`
	IconID     *int64    `json:"icon_id,omitempty"`
	StartTime  time.Time `json:"start_time"`
	PlannedEnd time.Time `json:"planned_end"`
}
