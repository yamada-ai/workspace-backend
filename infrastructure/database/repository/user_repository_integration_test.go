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

func TestUserRepository_Integration(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	testutil.CleanupTables(t, pool)

	queries := sqlc.New(pool)
	userRepository := repository.NewUserRepository(queries)
	ctx := context.Background()

	t.Run("Save and FindByName", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// Create a new user
		user, err := domain.NewUser("integration_test_user", 2, time.Now)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Save the user
		if err := userRepository.Save(ctx, user); err != nil {
			t.Fatalf("Failed to save user: %v", err)
		}

		// User should now have an ID
		if user.ID == 0 {
			t.Fatal("User ID should be set after save")
		}

		// Find the user by name
		found, err := userRepository.FindByName(ctx, "integration_test_user")
		if err != nil {
			t.Fatalf("Failed to find user by name: %v", err)
		}

		// Verify the user data
		if found.Name != "integration_test_user" {
			t.Errorf("Expected name 'integration_test_user', got %s", found.Name)
		}
		if found.Tier != 2 {
			t.Errorf("Expected tier 2, got %d", found.Tier)
		}
		if found.ID != user.ID {
			t.Errorf("Expected ID %d, got %d", user.ID, found.ID)
		}
	})

	t.Run("FindByID", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// Create and save a user
		user, _ := domain.NewUser("findbyid_test", 1, time.Now)
		if err := userRepository.Save(ctx, user); err != nil {
			t.Fatalf("Failed to save user: %v", err)
		}

		// Find by ID
		found, err := userRepository.FindByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to find user by ID: %v", err)
		}

		if found.Name != "findbyid_test" {
			t.Errorf("Expected name 'findbyid_test', got %s", found.Name)
		}
	})

	t.Run("FindByName_NotFound", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		_, err := userRepository.FindByName(ctx, "nonexistent_user")
		if err != domain.ErrUserNotFound {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		_, err := userRepository.FindByID(ctx, 99999)
		if err != domain.ErrUserNotFound {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("Save_DuplicateName", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// Create and save first user
		user1, _ := domain.NewUser("duplicate_test", 1, time.Now)
		if err := userRepository.Save(ctx, user1); err != nil {
			t.Fatalf("Failed to save first user: %v", err)
		}

		// Try to save second user with same name
		user2, _ := domain.NewUser("duplicate_test", 2, time.Now)
		err := userRepository.Save(ctx, user2)

		// Should return an error (unique constraint violation)
		if err == nil {
			t.Error("Expected error when saving duplicate user name, got nil")
		}
	})

	t.Run("Save_UpdateExistingUser", func(t *testing.T) {
		testutil.CleanupTables(t, pool)

		// Create and save a user
		user, _ := domain.NewUser("update_test", 1, time.Now)
		if err := userRepository.Save(ctx, user); err != nil {
			t.Fatalf("Failed to save user: %v", err)
		}

		originalID := user.ID

		// Modify the user
		user.Tier = 3

		// Save again (should update)
		if err := userRepository.Save(ctx, user); err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		// ID should remain the same
		if user.ID != originalID {
			t.Errorf("Expected ID to remain %d, got %d", originalID, user.ID)
		}

		// Verify the update
		found, err := userRepository.FindByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to find updated user: %v", err)
		}

		if found.Tier != 3 {
			t.Errorf("Expected tier 3, got %d", found.Tier)
		}
	})
}
