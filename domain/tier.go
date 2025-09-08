package domain

import (
	"fmt"
	"strings"
)

// Tier ユーザーのティア（フォロワー～ティア3）
type Tier int

const (
	TierUnknown Tier = iota
	Tier1            // 1
	Tier2            // 2
	Tier3            // 3
)

// 値を 1 始まりにしたい場合
func (t Tier) Int() int {
	switch t {
	case Tier1:
		return 1
	case Tier2:
		return 2
	case Tier3:
		return 3
	default:
		return 0
	}
}

// バリデーション
func (t Tier) Valid() bool {
	return t == Tier1 || t == Tier2 || t == Tier3
}

// 表示用
func (t Tier) String() string {
	switch t {
	case Tier1:
		return "Tier1"
	case Tier2:
		return "Tier2"
	case Tier3:
		return "Tier3"
	default:
		return "Unknown"
	}
}

// 文字列→Tier
func ParseTier(s string) (Tier, error) {
	s = strings.TrimSpace(s)
	us := strings.ToUpper(s)
	switch us {
	case "1", "TIER1":
		return Tier1, nil
	case "2", "TIER2":
		return Tier2, nil
	case "3", "TIER3":
		return Tier3, nil
	default:
		return TierUnknown, fmt.Errorf("invalid tier: %s", s)
	}
}
