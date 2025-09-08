package domain

import "testing"

func TestTier_Int_Values(t *testing.T) {
	if Tier1.Int() != 1 || Tier2.Int() != 2 || Tier3.Int() != 3 {
		t.Fatalf("Int() should map Tier1..3 -> 1..3")
	}
	if TierUnknown.Int() != 0 {
		t.Fatalf("Int() for unknown should be 0")
	}
}

func TestTier_Valid(t *testing.T) {
	if !Tier1.Valid() || !Tier2.Valid() || !Tier3.Valid() {
		t.Fatalf("Tier1..3 must be valid")
	}
	if TierUnknown.Valid() {
		t.Fatalf("TierUnknown must be invalid")
	}
	if Tier(999).Valid() {
		t.Fatalf("out-of-range tier must be invalid")
	}
}

func TestTier_String(t *testing.T) {
	if Tier1.String() != "Tier1" || Tier2.String() != "Tier2" || Tier3.String() != "Tier3" {
		t.Fatalf("String mismatch for Tier1..3")
	}
	if TierUnknown.String() != "Unknown" {
		t.Fatalf("String mismatch for Unknown")
	}
}

func TestParseTier(t *testing.T) {
	cases := []struct {
		in    string
		want  Tier
		isErr bool
	}{
		{"1", Tier1, false},
		{"2", Tier2, false},
		{"3", Tier3, false},
		{"Tier1", Tier1, false},
		{"tier2", Tier2, false},
		{"TIER3", Tier3, false},
		{"0", TierUnknown, true},
		{"x", TierUnknown, true},
		{"Tier0", TierUnknown, true},
	}

	for _, c := range cases {
		got, err := ParseTier(c.in)
		if c.isErr {
			if err == nil {
				t.Fatalf("ParseTier(%q) expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParseTier(%q) unexpected error: %v", c.in, err)
		}
		if got != c.want {
			t.Fatalf("ParseTier(%q) got %v want %v", c.in, got, c.want)
		}
	}
}
