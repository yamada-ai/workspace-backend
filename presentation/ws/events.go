package ws

import "time"

// EventType represents the type of WebSocket event
type EventType string

const (
	EventTypeSessionStart  EventType = "session_start"
	EventTypeSessionEnd    EventType = "session_end"
	EventTypeSessionExtend EventType = "session_extend"
)

// BaseEvent contains common fields for all events
type BaseEvent struct {
	Type EventType `json:"type"`
}

// SessionStartEvent is sent when a user starts a work session
type SessionStartEvent struct {
	Type        EventType `json:"type"`
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	UserName    string    `json:"user_name"`
	WorkName    string    `json:"work_name"`
	Tier        int       `json:"tier"`
	Icon        *string   `json:"icon,omitempty"`
	StartTime   time.Time `json:"start_time"`
	PlannedEnd  time.Time `json:"planned_end"`
}

// SessionEndEvent is sent when a user ends their work session
type SessionEndEvent struct {
	Type      EventType `json:"type"`
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ActualEnd time.Time `json:"actual_end"`
}

// SessionExtendEvent is sent when a user extends their session
type SessionExtendEvent struct {
	Type          EventType `json:"type"`
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	NewPlannedEnd time.Time `json:"new_planned_end"`
}

// Event is a union type of all possible WebSocket events
type Event interface {
	isEvent()
}

// Implement marker methods
func (SessionStartEvent) isEvent()  {}
func (SessionEndEvent) isEvent()    {}
func (SessionExtendEvent) isEvent() {}
