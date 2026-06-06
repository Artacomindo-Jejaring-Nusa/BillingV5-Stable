package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// InventoryItemType represents the categories/types of items.
type InventoryItemType struct {
	ID   uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
}

// TableName overrides the default table name for InventoryItemType
func (InventoryItemType) TableName() string {
	return "inventory_item_types"
}

// InventoryStatus represents the status of an item (e.g. available, assigned, broken).
type InventoryStatus struct {
	ID   uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
}

// TableName overrides the default table name for InventoryStatus
func (InventoryStatus) TableName() string {
	return "inventory_statuses"
}

// InventoryItem represents a physical item in the inventory.
type InventoryItem struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SerialNumber string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"serial_number"`
	MacAddress   *string    `gorm:"type:varchar(255);uniqueIndex" json:"mac_address"`
	Location     *string    `gorm:"type:varchar(255)" json:"location"`
	PurchaseDate *time.Time `gorm:"type:date" json:"purchase_date"`
	Notes        *string    `gorm:"type:text" json:"notes"`
	ItemTypeID   uint64     `gorm:"index;not null" json:"item_type_id"`
	StatusID     uint64     `gorm:"index;not null" json:"status_id"`
	PelangganID  *uint64    `gorm:"index" json:"pelanggan_id"`

	// Relationships
	ItemType           *InventoryItemType `gorm:"foreignKey:ItemTypeID" json:"item_type"`
	Status             *InventoryStatus   `gorm:"foreignKey:StatusID" json:"status"`
	Pelanggan          *Pelanggan         `gorm:"foreignKey:PelangganID" json:"pelanggan"`
	InventoryHistories []InventoryHistory `gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE" json:"inventory_histories"`
}

// TableName overrides the default table name for InventoryItem
func (InventoryItem) TableName() string {
	return "inventory_items"
}

// InventoryHistory tracks operations/changes on inventory items.
type InventoryHistory struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Action    string    `gorm:"type:varchar(255);not null" json:"action"`
	Timestamp time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;not null" json:"timestamp"`
	ItemID    uint64    `gorm:"index;not null" json:"item_id"`
	UserID    uint64    `gorm:"index;not null" json:"user_id"`

	// Relationships
	User          *User          `gorm:"foreignKey:UserID" json:"user"`
	InventoryItem *InventoryItem `gorm:"foreignKey:ItemID" json:"inventory_item"`
}

// TableName overrides the default table name for InventoryHistory
func (InventoryHistory) TableName() string {
	return "inventory_history"
}

// UnmarshalJSON custom unmarshaler for InventoryItem to handle date string unmarshaling for PurchaseDate.
func (i *InventoryItem) UnmarshalJSON(data []byte) error {
	type Alias InventoryItem
	aux := &struct {
		PurchaseDate *string `json:"purchase_date"`
		*Alias
	}{
		Alias: (*Alias)(i),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.PurchaseDate != nil && *aux.PurchaseDate != "" {
		t, err := time.Parse(time.RFC3339, *aux.PurchaseDate)
		if err == nil {
			i.PurchaseDate = &t
		} else {
			t2, err2 := time.Parse("2006-01-02", *aux.PurchaseDate)
			if err2 == nil {
				i.PurchaseDate = &t2
			} else {
				t3, err3 := time.Parse("2006-01-02 15:04:05", *aux.PurchaseDate)
				if err3 == nil {
					i.PurchaseDate = &t3
				} else {
					return fmt.Errorf("failed to parse purchase_date: %v", err2)
				}
			}
		}
	} else {
		i.PurchaseDate = nil
	}

	return nil
}

