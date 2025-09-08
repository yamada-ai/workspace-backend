package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyUserName = errors.New("user name must not be empty")
	ErrInvalidTier   = errors.New("invalid tier")
)

type User struct {
	ID        int64
	Name      string
	Tier      Tier
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(name string, tier Tier, now func() time.Time) (*User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyUserName
	}
	if !tier.Valid() {
		return nil, ErrInvalidTier
	}
	t := time.Now
	if now != nil {
		t = now
	}
	nowT := t()
	return &User{
		Name:      name,
		Tier:      tier,
		CreatedAt: nowT,
		UpdatedAt: nowT,
	}, nil
}

// 更新時刻を進める（更新操作後に呼ぶ）
func (u *User) Touch(now func() time.Time) {
	t := time.Now
	if now != nil {
		t = now
	}
	u.UpdatedAt = t()
}

// 追加のバリデーションが必要ならここに集約
func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return ErrEmptyUserName
	}
	if !u.Tier.Valid() {
		return ErrInvalidTier
	}
	return nil
}
