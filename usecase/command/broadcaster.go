package command

import "time"

// SessionStartBroadcast represents the data to broadcast when a session starts
type SessionStartBroadcast struct {
	SessionID  int64
	UserID     int64
	UserName   string
	WorkName   string
	Tier       int
	StartTime  time.Time
	PlannedEnd time.Time
}

// SessionEndBroadcast represents the data to broadcast when a session ends
type SessionEndBroadcast struct {
	SessionID int64
	UserID    int64
	ActualEnd time.Time
}

// WorkNameChangeBroadcast represents the data to broadcast when work name changes
type WorkNameChangeBroadcast struct {
	SessionID int64
	UserID    int64
	WorkName  string
}

// SessionExtendBroadcast represents the data to broadcast when a session is extended
type SessionExtendBroadcast struct {
	SessionID     int64
	UserID        int64
	NewPlannedEnd time.Time
}

// EventBroadcaster is an interface for broadcasting events to clients
type EventBroadcaster interface {
	BroadcastSessionStart(event SessionStartBroadcast)
	BroadcastSessionEnd(event SessionEndBroadcast)
	BroadcastWorkNameChange(event WorkNameChangeBroadcast)
	BroadcastSessionExtend(event SessionExtendBroadcast)
}

// NoOpBroadcaster is a no-op implementation of EventBroadcaster
// Useful for testing or when WebSocket is disabled
type NoOpBroadcaster struct{}

func (NoOpBroadcaster) BroadcastSessionStart(event SessionStartBroadcast)     {}
func (NoOpBroadcaster) BroadcastSessionEnd(event SessionEndBroadcast)         {}
func (NoOpBroadcaster) BroadcastWorkNameChange(event WorkNameChangeBroadcast) {}
func (NoOpBroadcaster) BroadcastSessionExtend(event SessionExtendBroadcast)   {}

// NoOpExpirationScheduler is a no-op implementation of ExpirationScheduler
// Useful for testing
type NoOpExpirationScheduler struct{}

func (NoOpExpirationScheduler) ScheduleExpiration(sessionID int64, userID int64, plannedEnd time.Time) {
}

// NoOpExpirationCanceller is a no-op implementation of ExpirationCanceller
// Useful for testing
type NoOpExpirationCanceller struct{}

func (NoOpExpirationCanceller) CancelExpiration(sessionID int64) {}
