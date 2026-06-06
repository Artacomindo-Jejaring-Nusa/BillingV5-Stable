package repository

import (
	"context"
	"errors"
	"strings"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) domain.InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) GetItems(ctx context.Context, limit, offset int, search string, itemTypeID, statusID, pelangganID *uint64) ([]domain.InventoryItem, int64, error) {
	var items []domain.InventoryItem
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.InventoryItem{})

	// Joins or Preloads
	if search != "" {
		db = db.Where("serial_number LIKE ? OR mac_address LIKE ? OR location LIKE ? OR notes LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if itemTypeID != nil {
		db = db.Where("item_type_id = ?", *itemTypeID)
	}

	if statusID != nil {
		db = db.Where("status_id = ?", *statusID)
	}

	if pelangganID != nil {
		db = db.Where("pelanggan_id = ?", *pelangganID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Preload associations
	db = db.Preload("ItemType").Preload("Status").Preload("Pelanggan")

	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset >= 0 {
		db = db.Offset(offset)
	}

	err := db.Order("id desc").Find(&items).Error
	return items, total, err
}

func (r *inventoryRepository) GetItemByID(ctx context.Context, id uint64) (*domain.InventoryItem, error) {
	var item domain.InventoryItem
	err := r.db.WithContext(ctx).
		Preload("ItemType").
		Preload("Status").
		Preload("Pelanggan").
		First(&item, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("inventory item not found")
		}
		return nil, err
	}
	return &item, nil
}

func (r *inventoryRepository) GetItemBySerialNumber(ctx context.Context, serialNumber string) (*domain.InventoryItem, error) {
	var item domain.InventoryItem
	err := r.db.WithContext(ctx).
		Where("serial_number = ?", serialNumber).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *inventoryRepository) GetItemByMacAddress(ctx context.Context, macAddress string) (*domain.InventoryItem, error) {
	var item domain.InventoryItem
	err := r.db.WithContext(ctx).
		Where("mac_address = ?", macAddress).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *inventoryRepository) CreateItem(ctx context.Context, item *domain.InventoryItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *inventoryRepository) UpdateItem(ctx context.Context, item *domain.InventoryItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *inventoryRepository) DeleteItem(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.InventoryItem{}, id).Error
}

func (r *inventoryRepository) GetItemTypes(ctx context.Context) ([]domain.InventoryItemType, error) {
	var types []domain.InventoryItemType
	err := r.db.WithContext(ctx).Order("name asc").Find(&types).Error
	return types, err
}

func (r *inventoryRepository) GetItemTypeByID(ctx context.Context, id uint64) (*domain.InventoryItemType, error) {
	var itemType domain.InventoryItemType
	err := r.db.WithContext(ctx).First(&itemType, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("item type not found")
		}
		return nil, err
	}
	return &itemType, nil
}

func (r *inventoryRepository) GetStatuses(ctx context.Context) ([]domain.InventoryStatus, error) {
	var statuses []domain.InventoryStatus
	err := r.db.WithContext(ctx).Order("name asc").Find(&statuses).Error
	return statuses, err
}

func (r *inventoryRepository) GetStatusByID(ctx context.Context, id uint64) (*domain.InventoryStatus, error) {
	var status domain.InventoryStatus
	err := r.db.WithContext(ctx).First(&status, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("status not found")
		}
		return nil, err
	}
	return &status, nil
}

func (r *inventoryRepository) GetStatusByName(ctx context.Context, name string) (*domain.InventoryStatus, error) {
	var status domain.InventoryStatus
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&status).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &status, nil
}

func (r *inventoryRepository) GetItemTypeByName(ctx context.Context, name string) (*domain.InventoryItemType, error) {
	var itemType domain.InventoryItemType
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&itemType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &itemType, nil
}

func (r *inventoryRepository) CreateItemType(ctx context.Context, itemType *domain.InventoryItemType) error {
	return r.db.WithContext(ctx).Create(itemType).Error
}

func (r *inventoryRepository) UpdateItemType(ctx context.Context, itemType *domain.InventoryItemType) error {
	return r.db.WithContext(ctx).Save(itemType).Error
}

func (r *inventoryRepository) DeleteItemType(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.InventoryItemType{}, id).Error
}

func (r *inventoryRepository) CreateStatus(ctx context.Context, status *domain.InventoryStatus) error {
	return r.db.WithContext(ctx).Create(status).Error
}

func (r *inventoryRepository) UpdateStatus(ctx context.Context, status *domain.InventoryStatus) error {
	return r.db.WithContext(ctx).Save(status).Error
}

func (r *inventoryRepository) DeleteStatus(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.InventoryStatus{}, id).Error
}

func (r *inventoryRepository) CreateHistory(ctx context.Context, history *domain.InventoryHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *inventoryRepository) GetHistoryByItemID(ctx context.Context, itemID uint64) ([]domain.InventoryHistory, error) {
	var history []domain.InventoryHistory
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("item_id = ?", itemID).
		Order("timestamp desc, id desc").
		Find(&history).Error
	return history, err
}

func (r *inventoryRepository) GetGlobalHistory(ctx context.Context) ([]domain.InventoryHistory, error) {
	var history []domain.InventoryHistory
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("InventoryItem").
		Order("timestamp desc, id desc").
		Find(&history).Error
	return history, err
}

func (r *inventoryRepository) GetAllSerialNumbersAndMacs(ctx context.Context) (map[string]bool, map[string]bool, error) {
	type SNMac struct {
		SerialNumber string  `gorm:"column:serial_number"`
		MacAddress   *string `gorm:"column:mac_address"`
	}
	var results []SNMac
	err := r.db.WithContext(ctx).Model(&domain.InventoryItem{}).Select("serial_number, mac_address").Find(&results).Error
	if err != nil {
		return nil, nil, err
	}

	sns := make(map[string]bool)
	macs := make(map[string]bool)
	for _, res := range results {
		if res.SerialNumber != "" {
			sns[strings.ToUpper(res.SerialNumber)] = true
		}
		if res.MacAddress != nil && *res.MacAddress != "" {
			macs[strings.ToUpper(*res.MacAddress)] = true
		}
	}
	return sns, macs, nil
}

