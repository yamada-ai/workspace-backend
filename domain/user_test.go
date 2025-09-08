package domain

import (
	"testing"
	"time"
)

func fixedNow() time.Time {
	return time.Date(2025, 9, 2, 12, 0, 0, 0, time.UTC)
}

func TestNewUser_Success(t *testing.T) {
	u, err := NewUser("  Alice  ", Tier1, fixedNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != 0 {
		t.Errorf("new user ID should be zero before persistence, got %d", u.ID)
	}
	if u.Name != "Alice" {
		t.Errorf("trim expected: 'Alice', got '%s'", u.Name)
	}
	if u.Tier != Tier1 {
		t.Errorf("tier mismatch")
	}
	if !u.CreatedAt.Equal(fixedNow()) || !u.UpdatedAt.Equal(fixedNow()) {
		t.Errorf("timestamps should be fixed to test time")
	}
	if err := u.Validate(); err != nil {
		t.Errorf("validate should pass, got %v", err)
	}
}

func TestNewUser_EmptyName(t *testing.T) {
	_, err := NewUser("   ", Tier2, fixedNow)
	if err == nil {
		t.Fatalf("expected error for empty name")
	}
	if err != ErrEmptyUserName {
		t.Fatalf("expected ErrEmptyUserName, got %v", err)
	}
}

func TestNewUser_InvalidTier(t *testing.T) {
	_, err := NewUser("Bob", Tier(999), fixedNow)
	if err == nil {
		t.Fatalf("expected error for invalid tier")
	}
	if err != ErrInvalidTier {
		t.Fatalf("expected ErrInvalidTier, got %v", err)
	}
}

func TestUser_Touch(t *testing.T) {
	u, _ := NewUser("Carol", Tier3, fixedNow)
	later := func() time.Time { return fixedNow().Add(5 * time.Minute) }
	u.Touch(later)
	if !u.UpdatedAt.Equal(later()) {
		t.Fatalf("updatedAt should be advanced by Touch")
	}
}
