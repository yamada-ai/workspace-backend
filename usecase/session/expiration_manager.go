package session

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/yamada-ai/workspace-backend/domain/repository"
)

// SessionExpirationManager manages automatic session expiration using timers
type SessionExpirationManager struct {
	timers          sync.Map // map[int64]*time.Timer
	sessionRepo     repository.SessionRepository
	completeService *CompleteSessionService
	now             func() time.Time
}

// NewSessionExpirationManager creates a new session expiration manager
func NewSessionExpirationManager(
	sessionRepo repository.SessionRepository,
	completeService *CompleteSessionService,
) *SessionExpirationManager {
	return &SessionExpirationManager{
		sessionRepo:     sessionRepo,
		completeService: completeService,
		now:             time.Now,
	}
}

// ScheduleExpiration schedules a session to be automatically completed at its planned end time
func (m *SessionExpirationManager) ScheduleExpiration(sessionID int64, userID int64, plannedEnd time.Time) {
	duration := time.Until(plannedEnd)

	// If already expired, don't schedule (should be handled elsewhere)
	if duration <= 0 {
		log.Printf("Session %d already expired, skipping timer", sessionID)
		return
	}

	timer := time.AfterFunc(duration, func() {
		m.handleExpiration(sessionID, userID)
	})

	m.timers.Store(sessionID, timer)
	log.Printf("Scheduled expiration for session %d in %v", sessionID, duration)
}

// CancelExpiration cancels a scheduled expiration (e.g., when user manually ends session)
func (m *SessionExpirationManager) CancelExpiration(sessionID int64) {
	if timerInterface, ok := m.timers.LoadAndDelete(sessionID); ok {
		if timer, ok := timerInterface.(*time.Timer); ok {
			timer.Stop()
			log.Printf("Cancelled expiration timer for session %d", sessionID)
		}
	}
}

// handleExpiration is called when a session reaches its planned end time
func (m *SessionExpirationManager) handleExpiration(sessionID int64, userID int64) {
	ctx := context.Background()

	log.Printf("Session %d reached planned end, completing automatically", sessionID)

	// 1. Fetch the session
	session, err := m.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		log.Printf("Failed to fetch session %d for expiration: %v", sessionID, err)
		return
	}

	// 2. Check if already completed (race condition with manual /out)
	if session.ActualEnd != nil {
		log.Printf("Session %d already completed, skipping auto-expiration", sessionID)
		return
	}

	// 3. Complete the session using the shared service
	if err := m.completeService.CompleteSession(ctx, session, userID); err != nil {
		log.Printf("Failed to auto-complete session %d: %v", sessionID, err)
		return
	}

	// 4. Remove timer from map
	m.timers.Delete(sessionID)

	log.Printf("Session %d auto-completed successfully", sessionID)
}

// InitializeFromDatabase loads all active sessions and schedules their expiration timers
// This is called on server startup to restore timers for existing sessions
func (m *SessionExpirationManager) InitializeFromDatabase(ctx context.Context) error {
	log.Println("Initializing session expiration timers from database...")

	// Fetch all active sessions
	sessions, err := m.sessionRepo.FindAllActive(ctx)
	if err != nil {
		return err
	}

	now := m.now()
	scheduledCount := 0
	expiredCount := 0

	for _, sessionInfo := range sessions {
		// If session is already expired, complete it immediately
		if sessionInfo.PlannedEnd.Before(now) || sessionInfo.PlannedEnd.Equal(now) {
			log.Printf("Session %d expired during downtime, completing now", sessionInfo.SessionID)

			// Fetch full session object
			session, err := m.sessionRepo.FindByID(ctx, sessionInfo.SessionID)
			if err != nil {
				log.Printf("Failed to fetch expired session %d: %v", sessionInfo.SessionID, err)
				continue
			}

			// Complete immediately
			if err := m.completeService.CompleteSession(ctx, session, sessionInfo.UserID); err != nil {
				log.Printf("Failed to complete expired session %d: %v", sessionInfo.SessionID, err)
				continue
			}

			expiredCount++
		} else {
			// Schedule future expiration
			m.ScheduleExpiration(sessionInfo.SessionID, sessionInfo.UserID, sessionInfo.PlannedEnd)
			scheduledCount++
		}
	}

	log.Printf("Initialized %d timers, completed %d expired sessions", scheduledCount, expiredCount)
	return nil
}
