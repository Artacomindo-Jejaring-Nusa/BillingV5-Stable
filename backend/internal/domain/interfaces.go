package domain

import (
	"context"
	"time"
)

// UserRepository defines the database operations for User entity
type UserRepository interface {
	GetByID(ctx context.Context, id uint64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context, limit, offset int, search string, roleID *uint64) ([]User, int64, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint64) error
	GetByResetToken(ctx context.Context, email, token string) (*User, error)
}

// UserUsecase defines the business logic operations for User entity
type UserUsecase interface {
	// Auth
	Authenticate(ctx context.Context, email, password string) (user *User, accessToken string, refreshToken string, expiresIn int, err error)
	RefreshToken(ctx context.Context, refreshTokenStr string) (newAccessToken string, newRefreshToken string, expiresIn int, err error)
	Logout(ctx context.Context, refreshTokenStr string) error
	LogoutAll(ctx context.Context, userID uint64) error

	// User CRUD
	GetProfile(ctx context.Context, id uint64) (*User, error)
	GetAll(ctx context.Context, page, pageSize int, search string, roleID *uint64) ([]User, int64, error)
	GetByID(ctx context.Context, id uint64) (*User, error)
	CreateUser(ctx context.Context, user *User, password string) (*User, error)
	UpdateUser(ctx context.Context, id uint64, updates map[string]interface{}) (*User, error)
	DeleteUser(ctx context.Context, id uint64) error

	// Password Management
	ForgotPassword(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, email, newPassword, token string) error
}

// TokenBlacklistRepository defines database operations for token blacklisting
type TokenBlacklistRepository interface {
	Blacklist(ctx context.Context, jti string, userID uint64, tokenType string, expiresAt time.Time, reason string) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
	BlacklistAllForUser(ctx context.Context, userID uint64) error
}

// RoleRepository defines database operations for Role
type RoleRepository interface {
	GetAll(ctx context.Context) ([]Role, error)
	GetByID(ctx context.Context, id uint64) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	Create(ctx context.Context, role *Role, permissionIDs []uint64) (*Role, error)
	Update(ctx context.Context, id uint64, name string, permissionIDs []uint64) (*Role, error)
	Delete(ctx context.Context, id uint64) error
}

// RoleUsecase defines business logic for Role management
type RoleUsecase interface {
	GetAll(ctx context.Context) ([]Role, error)
	GetByID(ctx context.Context, id uint64) (*Role, error)
	Create(ctx context.Context, name string, permissionIDs []uint64) (*Role, error)
	Update(ctx context.Context, id uint64, name string, permissionIDs []uint64) (*Role, error)
	Delete(ctx context.Context, id uint64) error
}

// PermissionRepository defines database operations for Permission
type PermissionRepository interface {
	GetAll(ctx context.Context) ([]Permission, error)
	GetByIDs(ctx context.Context, ids []uint64) ([]Permission, error)
}

// PermissionUsecase defines business logic for Permission
type PermissionUsecase interface {
	GetAll(ctx context.Context) ([]Permission, error)
}
