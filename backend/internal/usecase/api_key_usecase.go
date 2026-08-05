package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"billing-backend/internal/domain"
	"billing-backend/pkg/utils"
)

type apiKeyUsecase struct {
	apiKeyRepo domain.APIKeyRepository
	roleRepo   domain.RoleRepository
	systemRepo domain.SystemRepository
}

func NewAPIKeyUsecase(apiKeyRepo domain.APIKeyRepository, roleRepo domain.RoleRepository, systemRepo domain.SystemRepository) domain.APIKeyUsecase {
	return &apiKeyUsecase{
		apiKeyRepo: apiKeyRepo,
		roleRepo:   roleRepo,
		systemRepo: systemRepo,
	}
}

func (u *apiKeyUsecase) logActivity(ctx context.Context, action string, details string) {
	if u.systemRepo == nil {
		return
	}
	userID := utils.GetUserIDFromCtx(ctx)
	log := &domain.ActivityLog{
		UserID:    userID,
		Action:    action,
		Details:   &details,
		Timestamp: time.Now(),
	}
	_ = u.systemRepo.CreateActivityLog(ctx, log)
}

func (u *apiKeyUsecase) FetchAll(ctx context.Context) ([]domain.APIKey, error) {
	return u.apiKeyRepo.GetAll(ctx)
}

func (u *apiKeyUsecase) CreateAPIKey(ctx context.Context, name string, roleID uint64) (*domain.APIKey, string, error) {
	// 1. Verify role exists
	role, err := u.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, "", err
	}
	if role == nil {
		return nil, "", errors.New("role not found")
	}

	// 2. Generate raw key: jk_live_ + 32 hex chars
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	hexStr := hex.EncodeToString(randBytes)
	rawKey := fmt.Sprintf("jk_live_%s", hexStr) // total length: 8 + 32 = 40 chars

	// 3. Compute hash and prefix
	prefix := rawKey[:16] // jk_live_xxxxxxxx
	hashBytes := sha256.Sum256([]byte(rawKey))
	tokenHash := hex.EncodeToString(hashBytes[:])

	apiKey := &domain.APIKey{
		Name:      name,
		Prefix:    prefix,
		TokenHash: tokenHash,
		RoleID:    roleID,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 4. Save to DB
	if err := u.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, "", err
	}

	// Fetch with Role loaded for response
	apiKey.Role = *role

	// 5. Log activity
	u.logActivity(ctx, "CREATE_API_KEY", fmt.Sprintf("Created API Key '%s' (ID: %d, Prefix: %s, Role: %s)", name, apiKey.ID, prefix, role.Name))

	return apiKey, rawKey, nil
}

func (u *apiKeyUsecase) ToggleAPIKey(ctx context.Context, id uint64) (*domain.APIKey, error) {
	apiKey, err := u.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, errors.New("api key not found")
	}

	apiKey.IsActive = !apiKey.IsActive
	apiKey.UpdatedAt = time.Now()

	if err := u.apiKeyRepo.Update(ctx, apiKey); err != nil {
		return nil, err
	}

	action := "DEACTIVATE_API_KEY"
	if apiKey.IsActive {
		action = "ACTIVATE_API_KEY"
	}

	u.logActivity(ctx, action, fmt.Sprintf("Toggled API Key '%s' (ID: %d, Active: %t)", apiKey.Name, apiKey.ID, apiKey.IsActive))

	return apiKey, nil
}

func (u *apiKeyUsecase) DeleteAPIKey(ctx context.Context, id uint64) error {
	apiKey, err := u.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if apiKey == nil {
		return errors.New("api key not found")
	}

	if err := u.apiKeyRepo.Delete(ctx, id); err != nil {
		return err
	}

	u.logActivity(ctx, "DELETE_API_KEY", fmt.Sprintf("Deleted/Revoked API Key '%s' (ID: %d)", apiKey.Name, apiKey.ID))

	return nil
}
