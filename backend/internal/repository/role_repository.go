package repository

import (
	"context"
	"errors"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *gorm.DB) domain.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetAll(ctx context.Context) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.WithContext(ctx).Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetByID(ctx context.Context, id uint64) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Create(ctx context.Context, role *domain.Role, permissionIDs []uint64) (*domain.Role, error) {
	// Begin transaction
	tx := r.db.WithContext(ctx).Begin()

	if err := tx.Create(role).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Assign permissions if provided
	if len(permissionIDs) > 0 {
		var permissions []domain.Permission
		if err := tx.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := tx.Model(role).Association("Permissions").Append(&permissions); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	// Reload with permissions
	return r.GetByID(ctx, role.ID)
}

func (r *roleRepository) Update(ctx context.Context, id uint64, name string, permissionIDs []uint64) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	tx := r.db.WithContext(ctx).Begin()

	// Update name if provided
	if name != "" {
		role.Name = name
		if err := tx.Save(&role).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Update permissions (replace all)
	if permissionIDs != nil {
		var permissions []domain.Permission
		if len(permissionIDs) > 0 {
			if err := tx.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
		if err := tx.Model(&role).Association("Permissions").Replace(&permissions); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	// Reload with updated permissions
	return r.GetByID(ctx, id)
}

func (r *roleRepository) Delete(ctx context.Context, id uint64) error {
	// Remove permission associations first
	var role domain.Role
	if err := r.db.WithContext(ctx).First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return err
	}
	r.db.WithContext(ctx).Model(&role).Association("Permissions").Clear()
	return r.db.WithContext(ctx).Delete(&domain.Role{}, id).Error
}

// --- Permission Repository ---

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *gorm.DB) domain.PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) GetAll(ctx context.Context) ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.WithContext(ctx).Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) GetByIDs(ctx context.Context, ids []uint64) ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&permissions).Error
	return permissions, err
}
