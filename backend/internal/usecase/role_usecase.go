package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"billing-backend/internal/domain"
)

type roleUsecase struct {
	roleRepo domain.RoleRepository
}

// NewRoleUsecase creates a new role usecase
func NewRoleUsecase(roleRepo domain.RoleRepository) domain.RoleUsecase {
	return &roleUsecase{roleRepo: roleRepo}
}

func (u *roleUsecase) GetAll(ctx context.Context) ([]domain.Role, error) {
	return u.roleRepo.GetAll(ctx)
}

func (u *roleUsecase) GetByID(ctx context.Context, id uint64) (*domain.Role, error) {
	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role tidak ditemukan")
	}
	return role, nil
}

func (u *roleUsecase) Create(ctx context.Context, name string, permissionIDs []uint64) (*domain.Role, error) {
	// Check unique name
	existing, _ := u.roleRepo.GetByName(ctx, name)
	if existing != nil {
		return nil, fmt.Errorf("role dengan nama '%s' sudah ada", name)
	}

	role := &domain.Role{Name: name}
	return u.roleRepo.Create(ctx, role, permissionIDs)
}

func (u *roleUsecase) Update(ctx context.Context, id uint64, name string, permissionIDs []uint64) (*domain.Role, error) {
	existing, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("role tidak ditemukan")
	}

	return u.roleRepo.Update(ctx, id, name, permissionIDs)
}

func (u *roleUsecase) Delete(ctx context.Context, id uint64) error {
	existing, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("role tidak ditemukan")
	}
	return u.roleRepo.Delete(ctx, id)
}

// --- Permission Usecase ---

type permissionUsecase struct {
	permRepo domain.PermissionRepository
}

// NewPermissionUsecase creates a new permission usecase
func NewPermissionUsecase(permRepo domain.PermissionRepository) domain.PermissionUsecase {
	return &permissionUsecase{permRepo: permRepo}
}

func (u *permissionUsecase) GetAll(ctx context.Context) ([]domain.Permission, error) {
	return u.permRepo.GetAll(ctx)
}

func (u *permissionUsecase) Create(ctx context.Context, name string) (*domain.Permission, error) {
	clean := strings.TrimSpace(strings.ToLower(name))
	if clean == "" {
		return nil, errors.New("nama permission wajib diisi")
	}
	if _, err := u.permRepo.CreateIfNotExists(ctx, []string{clean}); err != nil {
		return nil, err
	}
	all, err := u.permRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].Name == clean {
			return &all[i], nil
		}
	}
	return nil, errors.New("gagal mengambil permission yang baru dibuat")
}

func (u *permissionUsecase) GenerateDefaultPermissions(ctx context.Context) (int, error) {
	defaults := []string{
		"view_customer_chat",
		"create_customer_chat",
		"edit_customer_chat",
		"delete_customer_chat",
	}
	return u.permRepo.CreateIfNotExists(ctx, defaults)
}
