package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/yamada-ai/workspace-backend/domain"
	"github.com/yamada-ai/workspace-backend/infrastructure/database/repository"
	"github.com/yamada-ai/workspace-backend/infrastructure/database/sqlc"
	"github.com/yamada-ai/workspace-backend/infrastructure/database/testutil"
)

func TestSessionRepository_Integration(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	testutil.CleanupTables(t, pool)

	queries := sqlc.New(pool)
	repo := repository.NewSessionRepository(queries)
	ctx := context.Background()

	// Helper to create a test user
	createTestUser := func(t *testing.T, name string, tier int32) int64 {
		t.Helper()
		return testutil.CreateTestUser(t, pool, name, tier)
	}

	t.Run("Save and FindByID", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		userID := createTestUser(t, "session_test_user", 1)

		// Create a new session
		session, err := domain.NewSession(userID, "テスト作業", 1*time.Hour, time.Now)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Save the session
		if err := repo.Save(ctx, session); err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		// Session should now have an ID
		if session.ID == 0 {
			t.Fatal("Session ID should be set after save")
		}

		// Find the session by ID
		found, err := repo.FindByID(ctx, session.ID)
		if err != nil {
			t.Fatalf("Failed to find session by ID: %v", err)
		}

		// Verify the session data
		if found.UserID != userID {
			t.Errorf("Expected user_id %d, got %d", userID, found.UserID)
		}
		if found.WorkName != "テスト作業" {
			t.Errorf("Expected work_name 'テスト作業', got %s", found.WorkName)
		}
		if found.ActualEnd != nil {
			t.Error("Expected actual_end to be nil")
		}
	})

	t.Run("FindActiveByUserID", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		userID := createTestUser(t, "active_session_user", 1)

		// Create an active session
		session, _ := domain.NewSession(userID, "アクティブセッション", 1*time.Hour, time.Now)
		if err := repo.Save(ctx, session); err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		// Find active session
		active, err := repo.FindActiveByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to find active session: %v", err)
		}

		if active.ID != session.ID {
			t.Errorf("Expected session ID %d, got %d", session.ID, active.ID)
		}
		if active.WorkName != "アクティブセッション" {
			t.Errorf("Expected work_name 'アクティブセッション', got %s", active.WorkName)
		}
	})

	t.Run("FindActiveByUserID_NoActiveSession", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		userID := createTestUser(t, "no_active_session_user", 1)

		// Try to find active session (should not exist)
		_, err := repo.FindActiveByUserID(ctx, userID)
		if err != domain.ErrSessionNotFound {
			t.Errorf("Expected ErrSessionNotFound, got %v", err)
		}
	})

	t.Run("FindActiveByUserID_OnlyCompletedSessions", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		userID := createTestUser(t, "completed_only_user", 1)

		// Create an active session first
		session, _ := domain.NewSession(userID, "完了済み", 1*time.Hour, time.Now)
		if err := repo.Save(ctx, session); err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		// Then complete it
		actualEnd := time.Now()
		session.ActualEnd = &actualEnd
		if err := repo.Save(ctx, session); err != nil {
			t.Fatalf("Failed to complete session: %v", err)
		}

		// Try to find active session (should not exist)
		_, err := repo.FindActiveByUserID(ctx, userID)
		if err != domain.ErrSessionNotFound {
			t.Errorf("Expected ErrSessionNotFound, got %v", err)
		}
	})

	t.Run("Save_UpdateExistingSession", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		userID := createTestUser(t, "update_session_user", 1)

		// Create and save a session
		session, _ := domain.NewSession(userID, "元の作業", 1*time.Hour, time.Now)
		if err := repo.Save(ctx, session); err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		originalID := session.ID

		// Complete the session
		completedTime := time.Now()
		session.ActualEnd = &completedTime

		// Save again (should update)
		if err := repo.Save(ctx, session); err != nil {
			t.Fatalf("Failed to update session: %v", err)
		}

		// ID should remain the same
		if session.ID != originalID {
			t.Errorf("Expected ID to remain %d, got %d", originalID, session.ID)
		}

		// Verify the update
		found, err := repo.FindByID(ctx, session.ID)
		if err != nil {
			t.Fatalf("Failed to find updated session: %v", err)
		}

		if found.ActualEnd == nil {
			t.Error("Expected actual_end to be set")
		}
	})

	t.Run("MultipleActiveSessions_ShouldNotExist", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		userID := createTestUser(t, "multi_session_user", 1)

		// Create first active session
		session1, _ := domain.NewSession(userID, "セッション1", 1*time.Hour, time.Now)
		if err := repo.Save(ctx, session1); err != nil {
			t.Fatalf("Failed to save first session: %v", err)
		}

		// Find active session (should be session1)
		active, err := repo.FindActiveByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to find active session: %v", err)
		}

		if active.ID != session1.ID {
			t.Errorf("Expected session ID %d, got %d", session1.ID, active.ID)
		}

		// Complete the first session
		completedTime := time.Now()
		session1.ActualEnd = &completedTime
		if err := repo.Save(ctx, session1); err != nil {
			t.Fatalf("Failed to update first session: %v", err)
		}

		// Create second active session
		session2, _ := domain.NewSession(userID, "セッション2", 1*time.Hour, time.Now)
		if err := repo.Save(ctx, session2); err != nil {
			t.Fatalf("Failed to save second session: %v", err)
		}

		// Find active session (should now be session2)
		active, err = repo.FindActiveByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to find active session: %v", err)
		}

		if active.ID != session2.ID {
			t.Errorf("Expected session ID %d, got %d", session2.ID, active.ID)
		}
		if active.WorkName != "セッション2" {
			t.Errorf("Expected work_name 'セッション2', got %s", active.WorkName)
		}
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		_, err := repo.FindByID(ctx, 99999)
		if err != domain.ErrSessionNotFound {
			t.Errorf("Expected ErrSessionNotFound, got %v", err)
		}
	})
}
