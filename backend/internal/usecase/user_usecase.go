package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"billing-backend/config"
	"billing-backend/internal/domain"
	"billing-backend/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo      domain.UserRepository
	tokenBlacklist domain.TokenBlacklistRepository
	config        *config.Config
}

// NewUserUsecase creates a new user usecase
func NewUserUsecase(
	userRepo domain.UserRepository,
	tokenBlacklist domain.TokenBlacklistRepository,
	cfg *config.Config,
) domain.UserUsecase {
	return &userUsecase{
		userRepo:      userRepo,
		tokenBlacklist: tokenBlacklist,
		config:        cfg,
	}
}

// Authenticate handles user login and returns access + refresh tokens.
// Mirrors Python routers/auth.py login_for_access_token
func (u *userUsecase) Authenticate(ctx context.Context, email, password string) (*domain.User, string, string, int, error) {
	// 1. Get user by email
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", 0, errors.New("email atau password salah")
	}
	if user == nil {
		return nil, "", "", 0, errors.New("email atau password salah")
	}

	// 2. Check if user is active
	if !user.IsActive {
		return nil, "", "", 0, errors.New("akun pengguna dinonaktifkan")
	}

	// 3. Compare password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", "", 0, errors.New("email atau password salah")
	}

	// 4. Determine Role Name
	roleName := "guest"
	if user.Role.ID != 0 && user.Role.Name != "" {
		roleName = user.Role.Name
	}

	// 5. Generate Access Token
	accessToken, err := utils.GenerateJWT(u.config, user.ID, user.Email, roleName)
	if err != nil {
		return nil, "", "", 0, errors.New("gagal membuat token autentikasi")
	}

	// 6. Generate Refresh Token
	refreshToken, _, err := utils.GenerateRefreshToken(u.config, user.ID, user.Email)
	if err != nil {
		return nil, "", "", 0, errors.New("gagal membuat refresh token")
	}

	expiresIn := u.config.AccessTokenExpireMinutes * 60 // in seconds

	log.Printf("User %s logged in successfully", user.Email)

	return user, accessToken, refreshToken, expiresIn, nil
}

// RefreshToken validates a refresh token and generates new access + refresh tokens.
// Mirrors Python routers/auth.py refresh_access_token + token_service.py
func (u *userUsecase) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, int, error) {
	// 1. Validate refresh token
	claims, err := utils.ValidateRefreshToken(u.config, refreshTokenStr)
	if err != nil {
		return "", "", 0, errors.New("refresh token tidak valid atau telah kadaluarsa")
	}

	// 2. Check if token is blacklisted
	if claims.ID != "" {
		blacklisted, err := u.tokenBlacklist.IsBlacklisted(ctx, claims.ID)
		if err != nil {
			return "", "", 0, errors.New("gagal memeriksa status token")
		}
		if blacklisted {
			return "", "", 0, errors.New("refresh token telah dicabut")
		}
	}

	// 3. Verify user still exists and is active
	user, err := u.userRepo.GetByID(ctx, claims.UserID)
	if err != nil || user == nil {
		return "", "", 0, errors.New("pengguna tidak ditemukan")
	}
	if !user.IsActive {
		return "", "", 0, errors.New("akun pengguna dinonaktifkan")
	}

	// 4. Blacklist old refresh token (token rotation)
	if claims.ID != "" {
		expTime := time.Now().Add(7 * 24 * time.Hour)
		if claims.ExpiresAt != nil {
			expTime = claims.ExpiresAt.Time
		}
		_ = u.tokenBlacklist.Blacklist(ctx, claims.ID, claims.UserID, "refresh", expTime, "Token rotation")
	}

	// 5. Generate new tokens
	roleName := "guest"
	if user.Role.ID != 0 && user.Role.Name != "" {
		roleName = user.Role.Name
	}

	newAccessToken, err := utils.GenerateJWT(u.config, user.ID, user.Email, roleName)
	if err != nil {
		return "", "", 0, errors.New("gagal membuat access token baru")
	}

	newRefreshToken, _, err := utils.GenerateRefreshToken(u.config, user.ID, user.Email)
	if err != nil {
		return "", "", 0, errors.New("gagal membuat refresh token baru")
	}

	expiresIn := u.config.AccessTokenExpireMinutes * 60
	return newAccessToken, newRefreshToken, expiresIn, nil
}

// Logout blacklists a single refresh token.
// Mirrors Python routers/auth.py logout
func (u *userUsecase) Logout(ctx context.Context, refreshTokenStr string) error {
	claims, err := utils.ValidateRefreshToken(u.config, refreshTokenStr)
	if err != nil {
		return errors.New("refresh token tidak valid")
	}

	expTime := time.Now().Add(7 * 24 * time.Hour)
	if claims.ExpiresAt != nil {
		expTime = claims.ExpiresAt.Time
	}

	return u.tokenBlacklist.Blacklist(ctx, claims.ID, claims.UserID, "refresh", expTime, "User logout")
}

// LogoutAll revokes all tokens for a user.
// Mirrors Python routers/auth.py logout_all
func (u *userUsecase) LogoutAll(ctx context.Context, userID uint64) error {
	return u.tokenBlacklist.BlacklistAllForUser(ctx, userID)
}

// GetProfile returns the user profile by ID
func (u *userUsecase) GetProfile(ctx context.Context, id uint64) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	return user, nil
}

// GetAll returns all users with pagination and filtering.
// Mirrors Python routers/user.py read_all_users
func (u *userUsecase) GetAll(ctx context.Context, page, pageSize int, search string, roleID *uint64) ([]domain.User, int64, error) {
	offset := (page - 1) * pageSize
	return u.userRepo.GetAll(ctx, pageSize, offset, search, roleID)
}

// GetByID returns a single user by ID
func (u *userUsecase) GetByID(ctx context.Context, id uint64) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	return user, nil
}

// CreateUser creates a new user with password validation.
// Mirrors Python routers/user.py create_user
func (u *userUsecase) CreateUser(ctx context.Context, user *domain.User, password string) (*domain.User, error) {
	// 1. Validate password strength
	isValid, validationErrors := utils.ValidatePasswordStrength(password)
	if !isValid {
		return nil, fmt.Errorf("password tidak memenuhi keamanan requirements: %v", validationErrors)
	}

	// 2. Check if email already exists
	existing, _ := u.userRepo.GetByEmail(ctx, user.Email)
	if existing != nil {
		return nil, fmt.Errorf("user dengan email '%s' sudah ada", user.Email)
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal memproses password")
	}
	user.Password = string(hashedPassword)

	// 4. Set password_changed_at
	now := time.Now()
	user.PasswordChangedAt = &now

	// 5. Create user
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 6. Reload user with relations
	return u.userRepo.GetByID(ctx, user.ID)
}

// UpdateUser updates user fields.
// Mirrors Python routers/user.py update_user
func (u *userUsecase) UpdateUser(ctx context.Context, id uint64, updates map[string]interface{}) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil || user == nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}

	// Handle password update with validation
	if pwd, ok := updates["password"]; ok {
		passwordStr, isString := pwd.(string)
		if isString && passwordStr != "" {
			isValid, validationErrors := utils.ValidatePasswordStrength(passwordStr)
			if !isValid {
				return nil, fmt.Errorf("password baru tidak memenuhi keamanan requirements: %v", validationErrors)
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordStr), bcrypt.DefaultCost)
			if err != nil {
				return nil, errors.New("gagal memproses password")
			}
			user.Password = string(hashedPassword)
			now := time.Now()
			user.PasswordChangedAt = &now
		}
		delete(updates, "password")
	}

	// Apply other field updates
	if name, ok := updates["name"]; ok {
		if n, isStr := name.(string); isStr {
			user.Name = n
		}
	}
	if email, ok := updates["email"]; ok {
		if e, isStr := email.(string); isStr {
			user.Email = e
		}
	}
	if roleID, ok := updates["role_id"]; ok {
		switch v := roleID.(type) {
		case float64:
			rid := uint64(v)
			user.RoleID = &rid
		case uint64:
			user.RoleID = &v
		case nil:
			user.RoleID = nil
		}
	}
	if isActive, ok := updates["is_active"]; ok {
		if active, isBool := isActive.(bool); isBool {
			user.IsActive = active
		}
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return u.userRepo.GetByID(ctx, user.ID)
}

// DeleteUser deletes a user by ID.
// Mirrors Python routers/user.py delete_user
func (u *userUsecase) DeleteUser(ctx context.Context, id uint64) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil || user == nil {
		return errors.New("pengguna tidak ditemukan")
	}
	return u.userRepo.Delete(ctx, id)
}

// ForgotPassword generates a reset token for password recovery.
// Mirrors Python routers/user.py forgot_password
func (u *userUsecase) ForgotPassword(ctx context.Context, email string) (string, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return "", errors.New("user dengan email tersebut tidak ditemukan")
	}

	// Generate reset token
	resetToken := uuid.New().String()
	expires := time.Now().Add(1 * time.Hour) // 1 hour validity

	user.ResetToken = &resetToken
	user.ResetTokenExpires = &expires

	if err := u.userRepo.Update(ctx, user); err != nil {
		return "", errors.New("gagal membuat token reset")
	}

	return resetToken, nil
}

// ResetPassword resets user password using a valid reset token.
// Mirrors Python routers/user.py reset_password
func (u *userUsecase) ResetPassword(ctx context.Context, email, newPassword, token string) error {
	user, err := u.userRepo.GetByResetToken(ctx, email, token)
	if err != nil || user == nil {
		return errors.New("email atau token reset tidak valid")
	}

	// Check token expiration
	if user.ResetTokenExpires != nil && user.ResetTokenExpires.Before(time.Now()) {
		return errors.New("token reset telah kadaluarsa")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal memproses password baru")
	}

	user.Password = string(hashedPassword)
	user.ResetToken = nil
	user.ResetTokenExpires = nil
	now := time.Now()
	user.PasswordChangedAt = &now

	return u.userRepo.Update(ctx, user)
}
