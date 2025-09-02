package domain

import (
	"errors"
	"strings"
	"time"
)

// 役割の定義（必要に応じて拡張）
type Role int

const (
	RoleUnknown Role = iota
	RoleUser
	RoleAdmin
)

func (r Role) Valid() bool {
	return r == RoleUser || r == RoleAdmin
}

var (
	ErrEmptyUserName = errors.New("user name must not be empty")
	ErrInvalidRole   = errors.New("invalid role")
)

// User 集約（最小）
type User struct {
	ID        int64
	Name      string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ファクトリ：不変条件を満たした User を作る
func NewUser(name string, role Role, now func() time.Time) (*User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyUserName
	}
	if !role.Valid() {
		return nil, ErrInvalidRole
	}
	t := time.Now
	if now != nil {
		t = now
	}
	nowT := t()
	return &User{
		Name:      name,
		Role:      role,
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
	if !u.Role.Valid() {
		return ErrInvalidRole
	}
	return nil
}
