package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yamada-ai/workspace-backend/infrastructure/database"
)

// SetupTestDB creates a test database connection and runs migrations
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Skip integration tests when running with -short flag
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Get database URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/workspace_test?sslmode=disable"
	}

	ctx := context.Background()

	// Run migrations (find migrations directory from project root)
	// Get the absolute path to migrations directory
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../../..")
	migrationsPath := filepath.Join(projectRoot, "migrations")

	if err := database.RunMigrations(dbURL, migrationsPath); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create connection pool
	pool, err := database.NewPool(ctx, dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up function to close pool after test
	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

// CleanupTables truncates all tables for a clean test state
func CleanupTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	queries := []string{
		"TRUNCATE TABLE sessions CASCADE",
		"TRUNCATE TABLE users CASCADE",
	}

	for _, query := range queries {
		if _, err := pool.Exec(ctx, query); err != nil {
			t.Fatalf("Failed to cleanup table: %v", err)
		}
	}
}

// CreateTestUser creates a test user and returns the ID
func CreateTestUser(t *testing.T, pool *pgxpool.Pool, name string, tier int32) int64 {
	t.Helper()

	ctx := context.Background()
	var id int64
	err := pool.QueryRow(ctx,
		"INSERT INTO users (name, tier) VALUES ($1, $2) RETURNING id",
		name, tier,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return id
}

// AssertUserExists checks if a user exists with the given name
func AssertUserExists(t *testing.T, pool *pgxpool.Pool, name string) int64 {
	t.Helper()

	ctx := context.Background()
	var id int64
	err := pool.QueryRow(ctx,
		"SELECT id FROM users WHERE name = $1",
		name,
	).Scan(&id)

	if err != nil {
		t.Fatalf("User %s does not exist: %v", name, err)
	}

	return id
}

// AssertSessionCount checks the number of sessions for a user
func AssertSessionCount(t *testing.T, pool *pgxpool.Pool, userID int64, expected int) {
	t.Helper()

	ctx := context.Background()
	var count int
	err := pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM sessions WHERE user_id = $1",
		userID,
	).Scan(&count)

	if err != nil {
		t.Fatalf("Failed to count sessions: %v", err)
	}

	if count != expected {
		t.Fatalf("Expected %d sessions, got %d", expected, count)
	}
}

// AssertActiveSessionExists checks if an active session exists for a user
func AssertActiveSessionExists(t *testing.T, pool *pgxpool.Pool, userID int64) int64 {
	t.Helper()

	ctx := context.Background()
	var sessionID int64
	err := pool.QueryRow(ctx,
		"SELECT id FROM sessions WHERE user_id = $1 AND actual_end IS NULL",
		userID,
	).Scan(&sessionID)

	if err != nil {
		t.Fatalf("No active session found for user %d: %v", userID, err)
	}

	return sessionID
}

// PrintDebugInfo prints database state for debugging
func PrintDebugInfo(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()

	// Print users
	rows, err := pool.Query(ctx, "SELECT id, name, tier FROM users")
	if err != nil {
		t.Logf("Failed to query users: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n=== Users ===")
	for rows.Next() {
		var id int64
		var name string
		var tier int32
		if err := rows.Scan(&id, &name, &tier); err != nil {
			t.Logf("Failed to scan user: %v", err)
			continue
		}
		fmt.Printf("ID: %d, Name: %s, Tier: %d\n", id, name, tier)
	}

	// Print sessions
	rows, err = pool.Query(ctx, "SELECT id, user_id, work_name, actual_end IS NULL as active FROM sessions")
	if err != nil {
		t.Logf("Failed to query sessions: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n=== Sessions ===")
	for rows.Next() {
		var id int64
		var userID int64
		var workName *string
		var active bool
		if err := rows.Scan(&id, &userID, &workName, &active); err != nil {
			t.Logf("Failed to scan session: %v", err)
			continue
		}
		workNameStr := "<nil>"
		if workName != nil {
			workNameStr = *workName
		}
		fmt.Printf("ID: %d, UserID: %d, WorkName: %s, Active: %v\n", id, userID, workNameStr, active)
	}
	fmt.Println()
}
