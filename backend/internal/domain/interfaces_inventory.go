package domain

import (
	"context"
)

type InventoryRepository interface {
	GetItems(ctx context.Context, limit, offset int, search string, itemTypeID, statusID, pelangganID *uint64) ([]InventoryItem, int64, error)
	GetItemByID(ctx context.Context, id uint64) (*InventoryItem, error)
	GetItemBySerialNumber(ctx context.Context, serialNumber string) (*InventoryItem, error)
	GetItemByMacAddress(ctx context.Context, macAddress string) (*InventoryItem, error)
	CreateItem(ctx context.Context, item *InventoryItem) error
	UpdateItem(ctx context.Context, item *InventoryItem) error
	DeleteItem(ctx context.Context, id uint64) error
	GetAllSerialNumbersAndMacs(ctx context.Context) (map[string]bool, map[string]bool, error)

	// Types & Statuses
	GetItemTypes(ctx context.Context) ([]InventoryItemType, error)
	GetItemTypeByID(ctx context.Context, id uint64) (*InventoryItemType, error)
	GetItemTypeByName(ctx context.Context, name string) (*InventoryItemType, error)
	CreateItemType(ctx context.Context, itemType *InventoryItemType) error
	UpdateItemType(ctx context.Context, itemType *InventoryItemType) error
	DeleteItemType(ctx context.Context, id uint64) error
	GetStatuses(ctx context.Context) ([]InventoryStatus, error)
	GetStatusByID(ctx context.Context, id uint64) (*InventoryStatus, error)
	GetStatusByName(ctx context.Context, name string) (*InventoryStatus, error)
	CreateStatus(ctx context.Context, status *InventoryStatus) error
	UpdateStatus(ctx context.Context, status *InventoryStatus) error
	DeleteStatus(ctx context.Context, id uint64) error

	// History
	CreateHistory(ctx context.Context, history *InventoryHistory) error
	GetHistoryByItemID(ctx context.Context, itemID uint64) ([]InventoryHistory, error)
	GetGlobalHistory(ctx context.Context) ([]InventoryHistory, error)
}

type InventoryUsecase interface {
	FetchItems(ctx context.Context, page, pageSize int, search string, itemTypeID, statusID, pelangganID *uint64) ([]InventoryItem, int64, error)
	GetItemByID(ctx context.Context, id uint64) (*InventoryItem, error)
	CreateItem(ctx context.Context, item *InventoryItem, userID uint64) (*InventoryItem, error)
	UpdateItem(ctx context.Context, id uint64, updates *InventoryItem, userID uint64) (*InventoryItem, error)
	DeleteItem(ctx context.Context, id uint64, userID uint64) error

	FetchItemTypes(ctx context.Context) ([]InventoryItemType, error)
	CreateItemType(ctx context.Context, itemType *InventoryItemType) (*InventoryItemType, error)
	UpdateItemType(ctx context.Context, id uint64, name string) (*InventoryItemType, error)
	DeleteItemType(ctx context.Context, id uint64) error

	FetchStatuses(ctx context.Context) ([]InventoryStatus, error)
	CreateStatus(ctx context.Context, status *InventoryStatus) (*InventoryStatus, error)
	UpdateStatus(ctx context.Context, id uint64, name string) (*InventoryStatus, error)
	DeleteStatus(ctx context.Context, id uint64) error

	AssignItem(ctx context.Context, itemID uint64, pelangganID uint64, notes string, userID uint64) error
	UnassignItem(ctx context.Context, itemID uint64, notes string, userID uint64) error
	FetchHistoryByItemID(ctx context.Context, itemID uint64) ([]InventoryHistory, error)
	FetchGlobalHistory(ctx context.Context) ([]InventoryHistory, error)
	BulkImport(ctx context.Context, items []InventoryItem, userID uint64) (int, int, []string, error)
}
