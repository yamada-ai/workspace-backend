package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/yamada-ai/workspace-backend/usecase/query"
	"github.com/yamada-ai/workspace-backend/usecase/session"
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
	userRepo := repository.NewUserRepositoryWithPool(pool)
	sessionRepo := repository.NewSessionRepository(sqlc.New(pool))
	completeService := session.NewCompleteSessionService(sessionRepo, command.NoOpBroadcaster{})
	expirationManager := session.NewSessionExpirationManager(sessionRepo, completeService)
	joinUseCase := command.NewJoinCommandUseCase(userRepo, sessionRepo, command.NoOpBroadcaster{}, expirationManager)
	outUseCase := command.NewOutCommandUseCase(userRepo, sessionRepo, completeService, expirationManager)
	moreUseCase := command.NewMoreCommandUseCase(userRepo, sessionRepo, expirationManager)
	getActiveSessionsUseCase := query.NewGetActiveSessionsUseCase(sessionRepo)
	getUserInfoUseCase := query.NewGetUserInfoUseCase(userRepo, sessionRepo)

	changeUseCase := command.NewChangeCommandUseCase(userRepo, sessionRepo, command.NoOpBroadcaster{})
	commandHandler := handler.NewCommandHandler(joinUseCase, outUseCase, moreUseCase, changeUseCase)
	queryHandler := handler.NewQueryHandler(getActiveSessionsUseCase, getUserInfoUseCase)
	unifiedHandler := handler.NewHandler(commandHandler, queryHandler)

	// Setup router
	r := chi.NewRouter()
	handlerFunc := dto.HandlerFromMux(unifiedHandler, r)

	// Create test server
	server := httptest.NewServer(handlerFunc)
	defer server.Close()

	t.Run("NewUser_JoinCommand", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// Prepare request
		reqBody := dto.JoinCommandRequest{
			UserName: "e2e_test_user",
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
		defer func() { _ = resp.Body.Close() }()

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
		defer func() { _ = resp.Body.Close() }()

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
		defer func() { _ = resp1.Body.Close() }()

		var response1 dto.JoinCommandResponse
		if err := json.NewDecoder(resp1.Body).Decode(&response1); err != nil {
			t.Fatalf("Failed to decode first response: %v", err)
		}

		// Second request: same user tries to join again
		reqBody2 := dto.JoinCommandRequest{
			UserName: "duplicate_join_user",
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
		defer func() { _ = resp2.Body.Close() }()

		// Check status code - should now return 409 Conflict
		if resp2.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", resp2.StatusCode)
		}

		// Parse error response
		var errorResp dto.ErrorResponse
		if err := json.NewDecoder(resp2.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode error response: %v", err)
		}

		// Verify error message
		if errorResp.Error == "" {
			t.Error("Expected error message to be set")
		}

		// Verify database state: should still have only 1 active session
		testutil.AssertSessionCount(t, pool, response1.UserId, 1)
	})

	t.Run("InvalidRequest_MissingUserName", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// Send request with missing user_name
		reqBody := dto.JoinCommandRequest{
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
		defer func() { _ = resp.Body.Close() }()

		// Should return 400 Bad Request
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("ConcurrentRequests_SameUser", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// Simulate concurrent requests from the same user with start barrier
		const numRequests = 5
		successResults := make(chan dto.JoinCommandResponse, numRequests)
		conflictResults := make(chan struct{}, numRequests)
		errors := make(chan error, numRequests)

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		// Start barrier to ensure truly concurrent execution
		start := make(chan struct{})

		for i := 0; i < numRequests; i++ {
			go func(workIndex int) {
				// Wait for start signal
				<-start

				reqBody := dto.JoinCommandRequest{
					UserName: "concurrent_user",
					WorkName: stringPtr("並行作業"),
				}
				bodyBytes, _ := json.Marshal(reqBody)

				resp, err := client.Post(
					server.URL+"/api/commands/join",
					"application/json",
					bytes.NewReader(bodyBytes),
				)
				if err != nil {
					errors <- err
					return
				}
				defer func() { _ = resp.Body.Close() }()

				// Accept both 200 OK (first request) and 409 Conflict (subsequent requests)
				if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
					errors <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
					return
				}

				// Collect successful responses (200 OK)
				if resp.StatusCode == http.StatusOK {
					var response dto.JoinCommandResponse
					if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
						errors <- err
						return
					}
					successResults <- response
				} else {
					// 409 Conflict
					conflictResults <- struct{}{}
				}
			}(i)
		}

		// Signal all goroutines to start simultaneously
		close(start)

		// Collect results
		responses := make([]dto.JoinCommandResponse, 0, numRequests)
		conflictCount := 0
		errorCount := 0
		for i := 0; i < numRequests; i++ {
			select {
			case resp := <-successResults:
				responses = append(responses, resp)
			case <-conflictResults:
				conflictCount++
			case err := <-errors:
				errorCount++
				t.Errorf("Unexpected error: %v", err)
			case <-time.After(10 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}

		// Verify we got exactly 1 successful response (first request wins)
		if len(responses) != 1 {
			t.Errorf("Expected exactly 1 successful response, got %d", len(responses))
		}

		// Verify remaining requests got 409 Conflict
		if conflictCount != numRequests-1 {
			t.Errorf("Expected %d conflict responses, got %d", numRequests-1, conflictCount)
		}

		// Verify no unexpected errors
		if errorCount > 0 {
			t.Errorf("Got %d unexpected errors", errorCount)
		}

		// Verify session_id from successful response
		if len(responses) > 0 {
			sessionID := responses[0].SessionId
			if sessionID == 0 {
				t.Error("Expected session_id to be set")
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
	userRepo := repository.NewUserRepositoryWithPool(pool)
	sessionRepo := repository.NewSessionRepository(sqlc.New(pool))
	completeService := session.NewCompleteSessionService(sessionRepo, command.NoOpBroadcaster{})
	expirationManager := session.NewSessionExpirationManager(sessionRepo, completeService)
	joinUseCase := command.NewJoinCommandUseCase(userRepo, sessionRepo, command.NoOpBroadcaster{}, expirationManager)
	outUseCase := command.NewOutCommandUseCase(userRepo, sessionRepo, completeService, expirationManager)
	moreUseCase := command.NewMoreCommandUseCase(userRepo, sessionRepo, expirationManager)
	getActiveSessionsUseCase := query.NewGetActiveSessionsUseCase(sessionRepo)
	getUserInfoUseCase := query.NewGetUserInfoUseCase(userRepo, sessionRepo)

	changeUseCase := command.NewChangeCommandUseCase(userRepo, sessionRepo, command.NoOpBroadcaster{})
	commandHandler := handler.NewCommandHandler(joinUseCase, outUseCase, moreUseCase, changeUseCase)
	queryHandler := handler.NewQueryHandler(getActiveSessionsUseCase, getUserInfoUseCase)
	unifiedHandler := handler.NewHandler(commandHandler, queryHandler)

	// Setup router
	r := chi.NewRouter()
	handlerFunc := dto.HandlerFromMux(unifiedHandler, r)

	// Create test server
	server := httptest.NewServer(handlerFunc)
	defer server.Close()

	t.Run("CompleteWorkflow", func(t *testing.T) {
		testutil.CleanupTables(t, pool)
		ctx := context.Background()

		// Step 1: User joins for the first time
		reqBody1 := dto.JoinCommandRequest{
			UserName: "workflow_user",
			WorkName: stringPtr("プロジェクトA"),
		}
		bodyBytes1, _ := json.Marshal(reqBody1)

		resp1, _ := http.Post(server.URL+"/api/commands/join", "application/json", bytes.NewReader(bodyBytes1))
		var response1 dto.JoinCommandResponse
		if err := json.NewDecoder(resp1.Body).Decode(&response1); err != nil {
			t.Fatalf("Failed to decode first response: %v", err)
		}
		_ = resp1.Body.Close()

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
			WorkName: stringPtr("プロジェクトB"),
		}
		bodyBytes2, _ := json.Marshal(reqBody2)

		resp2, _ := http.Post(server.URL+"/api/commands/join", "application/json", bytes.NewReader(bodyBytes2))
		var response2 dto.JoinCommandResponse
		if err := json.NewDecoder(resp2.Body).Decode(&response2); err != nil {
			t.Fatalf("Failed to decode second response: %v", err)
		}
		_ = resp2.Body.Close()

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
