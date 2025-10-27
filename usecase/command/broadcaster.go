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

// EventBroadcaster is an interface for broadcasting events to clients
type EventBroadcaster interface {
	BroadcastSessionStart(event SessionStartBroadcast)
	BroadcastSessionEnd(event SessionEndBroadcast)
}

// NoOpBroadcaster is a no-op implementation of EventBroadcaster
// Useful for testing or when WebSocket is disabled
type NoOpBroadcaster struct{}

func (NoOpBroadcaster) BroadcastSessionStart(event SessionStartBroadcast) {}
func (NoOpBroadcaster) BroadcastSessionEnd(event SessionEndBroadcast)     {}
