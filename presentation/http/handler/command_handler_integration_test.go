package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yamada-ai/workspace-backend/infrastructure/database/repository"
	"github.com/yamada-ai/workspace-backend/infrastructure/database/sqlc"
	"github.com/yamada-ai/workspace-backend/infrastructure/database/testutil"
	"github.com/yamada-ai/workspace-backend/presentation/http/dto"
	"github.com/yamada-ai/workspace-backend/presentation/http/handler"
	"github.com/yamada-ai/workspace-backend/usecase/command"
)

func TestCommandHandler_JoinCommand_E2E(t *testing.T) {
	// Skip integration tests when running with -short flag
	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	// Setup test database
	pool := testutil.SetupTestDB(t)
	testutil.CleanupTables(t, pool)

	// Create dependencies
	queries := sqlc.New(pool)
	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	joinUseCase := command.NewJoinCommandUseCase(userRepo, sessionRepo)
	commandHandler := handler.NewCommandHandler(joinUseCase)

	// Setup router
	r := chi.NewRouter()
	handlerFunc := dto.HandlerFromMux(commandHandler, r)

	// Create test server
	server := httptest.NewServer(handlerFunc)
	defer server.Close()

	t.Run("NewUser_JoinCommand", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// Prepare request
		reqBody := dto.JoinCommandRequest{
			UserName: "e2e_test_user",
			Tier:     1,
			WorkName: stringPtr("E2Eテスト作業"),
		}
		bodyBytes, _ := json.Marshal(reqBody)

		// Send POST request
		resp, err := http.Post(
			server.URL+"/api/commands/join",
			"application/json",
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var response dto.JoinCommandResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Verify response
		if response.UserId == 0 {
			t.Error("Expected user_id to be set")
		}
		if response.SessionId == 0 {
			t.Error("Expected session_id to be set")
		}
		if response.WorkName == nil || *response.WorkName != "E2Eテスト作業" {
			t.Errorf("Expected work_name 'E2Eテスト作業', got %v", response.WorkName)
		}

		// Verify database state
		userID := testutil.AssertUserExists(t, pool, "e2e_test_user")
		if userID != response.UserId {
			t.Errorf("Expected user_id %d in DB, got %d", response.UserId, userID)
		}

		sessionID := testutil.AssertActiveSessionExists(t, pool, userID)
		if sessionID != response.SessionId {
			t.Errorf("Expected session_id %d in DB, got %d", response.SessionId, sessionID)
		}
	})

	t.Run("ExistingUser_NewSession", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// Create existing user with no active session
		testutil.CreateTestUser(t, pool, "existing_user", 2)

		// Send join request
		reqBody := dto.JoinCommandRequest{
			UserName: "existing_user",
			Tier:     2,
			WorkName: stringPtr("新しい作業"),
		}
		bodyBytes, _ := json.Marshal(reqBody)

		resp, err := http.Post(
			server.URL+"/api/commands/join",
			"application/json",
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var response dto.JoinCommandResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Verify response
		if response.UserId == 0 {
			t.Error("Expected user_id to be set")
		}
		if response.SessionId == 0 {
			t.Error("Expected session_id to be set")
		}

		// Verify database state
		testutil.AssertActiveSessionExists(t, pool, response.UserId)
	})

	t.Run("ExistingUser_AlreadyInSession", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// First request: create user and session
		reqBody1 := dto.JoinCommandRequest{
			UserName: "duplicate_join_user",
			Tier:     1,
			WorkName: stringPtr("最初の作業"),
		}
		bodyBytes1, _ := json.Marshal(reqBody1)

		resp1, err := http.Post(
			server.URL+"/api/commands/join",
			"application/json",
			bytes.NewReader(bodyBytes1),
		)
		if err != nil {
			t.Fatalf("Failed to send first request: %v", err)
		}
		defer resp1.Body.Close()

		var response1 dto.JoinCommandResponse
		json.NewDecoder(resp1.Body).Decode(&response1)

		// Second request: same user tries to join again
		reqBody2 := dto.JoinCommandRequest{
			UserName: "duplicate_join_user",
			Tier:     1,
			WorkName: stringPtr("二回目の作業"),
		}
		bodyBytes2, _ := json.Marshal(reqBody2)

		resp2, err := http.Post(
			server.URL+"/api/commands/join",
			"application/json",
			bytes.NewReader(bodyBytes2),
		)
		if err != nil {
			t.Fatalf("Failed to send second request: %v", err)
		}
		defer resp2.Body.Close()

		// Check status code
		if resp2.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp2.StatusCode)
		}

		// Parse response
		var response2 dto.JoinCommandResponse
		if err := json.NewDecoder(resp2.Body).Decode(&response2); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Verify that the same session is returned
		if response2.SessionId != response1.SessionId {
			t.Errorf("Expected same session_id %d, got %d", response1.SessionId, response2.SessionId)
		}

		// Work name should be from the first session, not the second
		if response2.WorkName == nil || *response2.WorkName != "最初の作業" {
			t.Errorf("Expected work_name '最初の作業', got %v", response2.WorkName)
		}

		// Verify database state: should still have only 1 active session
		testutil.AssertSessionCount(t, pool, response1.UserId, 1)
	})

	t.Run("InvalidRequest_MissingUserName", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// Send request with missing user_name
		reqBody := dto.JoinCommandRequest{
			Tier:     1,
			WorkName: stringPtr("作業"),
		}
		bodyBytes, _ := json.Marshal(reqBody)

		resp, err := http.Post(
			server.URL+"/api/commands/join",
			"application/json",
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		// Should return 400 Bad Request
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("InvalidRequest_InvalidTier", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// Send request with invalid tier
		reqBody := dto.JoinCommandRequest{
			UserName: "invalid_tier_user",
			Tier:     99, // Invalid tier
			WorkName: stringPtr("作業"),
		}
		bodyBytes, _ := json.Marshal(reqBody)

		resp, err := http.Post(
			server.URL+"/api/commands/join",
			"application/json",
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		// Should return 400 Bad Request
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("ConcurrentRequests_SameUser", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// Simulate concurrent requests from the same user
		const numRequests = 5
		results := make(chan dto.JoinCommandResponse, numRequests)
		errors := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(workIndex int) {
				reqBody := dto.JoinCommandRequest{
					UserName: "concurrent_user",
					Tier:     1,
					WorkName: stringPtr("並行作業"),
				}
				bodyBytes, _ := json.Marshal(reqBody)

				resp, err := http.Post(
					server.URL+"/api/commands/join",
					"application/json",
					bytes.NewReader(bodyBytes),
				)
				if err != nil {
					errors <- err
					return
				}
				defer resp.Body.Close()

				var response dto.JoinCommandResponse
				if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
					errors <- err
					return
				}

				results <- response
			}(i)
		}

		// Collect results
		responses := make([]dto.JoinCommandResponse, 0, numRequests)
		for i := 0; i < numRequests; i++ {
			select {
			case resp := <-results:
				responses = append(responses, resp)
			case err := <-errors:
				t.Errorf("Request failed: %v", err)
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}

		// All responses should have the same session_id
		firstSessionID := responses[0].SessionId
		for i, resp := range responses {
			if resp.SessionId != firstSessionID {
				t.Errorf("Request %d: Expected session_id %d, got %d", i, firstSessionID, resp.SessionId)
			}
		}

		// Verify database state: should have exactly 1 active session
		userID := testutil.AssertUserExists(t, pool, "concurrent_user")
		testutil.AssertSessionCount(t, pool, userID, 1)
	})
}

func TestCommandHandler_Integration_FullFlow(t *testing.T) {
	// Skip integration tests when running with -short flag
	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	// Setup test database
	pool := testutil.SetupTestDB(t)
	testutil.CleanupTables(t, pool)

	// Create dependencies
	queries := sqlc.New(pool)
	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	joinUseCase := command.NewJoinCommandUseCase(userRepo, sessionRepo)
	commandHandler := handler.NewCommandHandler(joinUseCase)

	// Setup router
	r := chi.NewRouter()
	handlerFunc := dto.HandlerFromMux(commandHandler, r)

	// Create test server
	server := httptest.NewServer(handlerFunc)
	defer server.Close()

	t.Run("CompleteWorkflow", func(t *testing.T) {
		testutil.CleanupTables(t, pool)
		ctx := context.Background()

		// Step 1: User joins for the first time
		reqBody1 := dto.JoinCommandRequest{
			UserName: "workflow_user",
			Tier:     2,
			WorkName: stringPtr("プロジェクトA"),
		}
		bodyBytes1, _ := json.Marshal(reqBody1)

		resp1, _ := http.Post(server.URL+"/api/commands/join", "application/json", bytes.NewReader(bodyBytes1))
		var response1 dto.JoinCommandResponse
		json.NewDecoder(resp1.Body).Decode(&response1)
		resp1.Body.Close()

		// Verify initial state
		if response1.SessionId == 0 {
			t.Fatal("Expected session_id to be set")
		}

		// Step 2: Complete the session manually in DB (simulating /out command)
		completedTime := time.Now()
		_, err := pool.Exec(ctx, "UPDATE sessions SET actual_end = $1 WHERE id = $2", completedTime, response1.SessionId)
		if err != nil {
			t.Fatalf("Failed to complete session: %v", err)
		}

		// Step 3: User joins again with a different task
		reqBody2 := dto.JoinCommandRequest{
			UserName: "workflow_user",
			Tier:     2,
			WorkName: stringPtr("プロジェクトB"),
		}
		bodyBytes2, _ := json.Marshal(reqBody2)

		resp2, _ := http.Post(server.URL+"/api/commands/join", "application/json", bytes.NewReader(bodyBytes2))
		var response2 dto.JoinCommandResponse
		json.NewDecoder(resp2.Body).Decode(&response2)
		resp2.Body.Close()

		// Verify new session was created
		if response2.SessionId == response1.SessionId {
			t.Error("Expected a new session_id, got the same one")
		}
		if response2.WorkName == nil || *response2.WorkName != "プロジェクトB" {
			t.Errorf("Expected work_name 'プロジェクトB', got %v", response2.WorkName)
		}

		// Verify database state: should have 2 sessions (1 completed, 1 active)
		testutil.AssertSessionCount(t, pool, response1.UserId, 2)
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
