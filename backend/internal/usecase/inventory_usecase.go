package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"billing-backend/internal/domain"
)

type inventoryUsecase struct {
	repo          domain.InventoryRepository
	pelangganRepo domain.PelangganRepository
	systemRepo    domain.SystemRepository
}

func NewInventoryUsecase(r domain.InventoryRepository, pr domain.PelangganRepository, sr domain.SystemRepository) domain.InventoryUsecase {
	return &inventoryUsecase{
		repo:          r,
		pelangganRepo: pr,
		systemRepo:    sr,
	}
}

func (u *inventoryUsecase) logActivity(ctx context.Context, userID uint64, action string, details string) {
	log := &domain.ActivityLog{
		UserID:    userID,
		Action:    action,
		Details:   &details,
		Timestamp: time.Now(),
	}
	_ = u.systemRepo.CreateActivityLog(ctx, log)
}

func (u *inventoryUsecase) FetchItems(ctx context.Context, page, pageSize int, search string, itemTypeID, statusID, pelangganID *uint64) ([]domain.InventoryItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return u.repo.GetItems(ctx, pageSize, offset, search, itemTypeID, statusID, pelangganID)
}

func (u *inventoryUsecase) GetItemByID(ctx context.Context, id uint64) (*domain.InventoryItem, error) {
	return u.repo.GetItemByID(ctx, id)
}

func (u *inventoryUsecase) CreateItem(ctx context.Context, item *domain.InventoryItem, userID uint64) (*domain.InventoryItem, error) {
	// Validate Serial Number
	if item.SerialNumber == "" {
		return nil, errors.New("serial number is required")
	}

	existingSN, err := u.repo.GetItemBySerialNumber(ctx, item.SerialNumber)
	if err != nil {
		return nil, err
	}
	if existingSN != nil {
		return nil, errors.New("serial number already exists")
	}

	// Validate MAC Address if provided
	if item.MacAddress != nil && *item.MacAddress != "" {
		existingMac, err := u.repo.GetItemByMacAddress(ctx, *item.MacAddress)
		if err != nil {
			return nil, err
		}
		if existingMac != nil {
			return nil, errors.New("mac address already exists")
		}
	}

	// Verify Item Type
	_, err = u.repo.GetItemTypeByID(ctx, item.ItemTypeID)
	if err != nil {
		return nil, fmt.Errorf("invalid item type: %v", err)
	}

	// Verify Status
	_, err = u.repo.GetStatusByID(ctx, item.StatusID)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %v", err)
	}

	// Save
	err = u.repo.CreateItem(ctx, item)
	if err != nil {
		return nil, err
	}

	// Fetch complete item (with preloaded associations)
	savedItem, err := u.repo.GetItemByID(ctx, item.ID)
	if err != nil {
		return item, nil // return partial if load fails
	}

	// Log History
	history := &domain.InventoryHistory{
		ItemID:    savedItem.ID,
		UserID:    userID,
		Action:    "Item Created",
		Timestamp: time.Now(),
	}
	_ = u.repo.CreateHistory(ctx, history)

	// Log Activity
	u.logActivity(ctx, userID, "Create Inventory Item", fmt.Sprintf("Created item with SN: %s, Type ID: %d", savedItem.SerialNumber, savedItem.ItemTypeID))

	return savedItem, nil
}

func (u *inventoryUsecase) UpdateItem(ctx context.Context, id uint64, updates *domain.InventoryItem, userID uint64) (*domain.InventoryItem, error) {
	existing, err := u.repo.GetItemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate Serial Number uniqueness if changed
	if updates.SerialNumber != "" && updates.SerialNumber != existing.SerialNumber {
		existingSN, err := u.repo.GetItemBySerialNumber(ctx, updates.SerialNumber)
		if err != nil {
			return nil, err
		}
		if existingSN != nil {
			return nil, errors.New("serial number already exists")
		}
		existing.SerialNumber = updates.SerialNumber
	}

	// Validate MAC Address uniqueness if changed
	if updates.MacAddress != nil && *updates.MacAddress != "" {
		if existing.MacAddress == nil || *updates.MacAddress != *existing.MacAddress {
			existingMac, err := u.repo.GetItemByMacAddress(ctx, *updates.MacAddress)
			if err != nil {
				return nil, err
			}
			if existingMac != nil {
				return nil, errors.New("mac address already exists")
			}
			existing.MacAddress = updates.MacAddress
		}
	} else if updates.MacAddress == nil {
		existing.MacAddress = nil
	}

	// Verify Item Type if changed
	if updates.ItemTypeID != 0 && updates.ItemTypeID != existing.ItemTypeID {
		_, err = u.repo.GetItemTypeByID(ctx, updates.ItemTypeID)
		if err != nil {
			return nil, fmt.Errorf("invalid item type: %v", err)
		}
		existing.ItemTypeID = updates.ItemTypeID
	}

	// Verify Status if changed
	if updates.StatusID != 0 && updates.StatusID != existing.StatusID {
		_, err = u.repo.GetStatusByID(ctx, updates.StatusID)
		if err != nil {
			return nil, fmt.Errorf("invalid status: %v", err)
		}
		existing.StatusID = updates.StatusID
	}

	existing.Location = updates.Location
	existing.PurchaseDate = updates.PurchaseDate
	existing.Notes = updates.Notes

	err = u.repo.UpdateItem(ctx, existing)
	if err != nil {
		return nil, err
	}

	// Refresh preloaded fields
	updatedItem, err := u.repo.GetItemByID(ctx, existing.ID)
	if err != nil {
		return existing, nil
	}

	// Log History
	history := &domain.InventoryHistory{
		ItemID:    updatedItem.ID,
		UserID:    userID,
		Action:    "Item Updated",
		Timestamp: time.Now(),
	}
	_ = u.repo.CreateHistory(ctx, history)

	// Log Activity
	u.logActivity(ctx, userID, "Update Inventory Item", fmt.Sprintf("Updated item ID: %d, SN: %s", updatedItem.ID, updatedItem.SerialNumber))

	return updatedItem, nil
}

func (u *inventoryUsecase) DeleteItem(ctx context.Context, id uint64, userID uint64) error {
	existing, err := u.repo.GetItemByID(ctx, id)
	if err != nil {
		return err
	}

	err = u.repo.DeleteItem(ctx, existing.ID)
	if err != nil {
		return err
	}

	// Log Activity
	u.logActivity(ctx, userID, "Delete Inventory Item", fmt.Sprintf("Deleted item ID: %d, SN: %s", existing.ID, existing.SerialNumber))

	return nil
}

func (u *inventoryUsecase) FetchItemTypes(ctx context.Context) ([]domain.InventoryItemType, error) {
	return u.repo.GetItemTypes(ctx)
}

func (u *inventoryUsecase) FetchStatuses(ctx context.Context) ([]domain.InventoryStatus, error) {
	return u.repo.GetStatuses(ctx)
}

func (u *inventoryUsecase) AssignItem(ctx context.Context, itemID uint64, pelangganID uint64, notes string, userID uint64) error {
	item, err := u.repo.GetItemByID(ctx, itemID)
	if err != nil {
		return err
	}

	pelanggan, err := u.pelangganRepo.GetByID(ctx, pelangganID)
	if err != nil {
		return fmt.Errorf("pelanggan not found: %v", err)
	}

	// Check if item status is "available"
	if item.Status != nil && item.Status.Name != "available" {
		return fmt.Errorf("item is not available (current status: %s)", item.Status.Name)
	}

	// Get assigned status ID
	assignedStatus, err := u.repo.GetStatusByName(ctx, "assigned")
	if err != nil || assignedStatus == nil {
		return errors.New("assigned status configuration not found in database")
	}

	// Update item
	item.StatusID = assignedStatus.ID
	item.PelangganID = &pelangganID
	if notes != "" {
		if item.Notes != nil && *item.Notes != "" {
			newNotes := fmt.Sprintf("%s\nAssigned note: %s", *item.Notes, notes)
			item.Notes = &newNotes
		} else {
			item.Notes = &notes
		}
	}

	err = u.repo.UpdateItem(ctx, item)
	if err != nil {
		return err
	}

	// Log History
	actionText := fmt.Sprintf("Assigned to pelanggan %s (ID: %d)", pelanggan.Nama, pelanggan.ID)
	history := &domain.InventoryHistory{
		ItemID:    item.ID,
		UserID:    userID,
		Action:    actionText,
		Timestamp: time.Now(),
	}
	_ = u.repo.CreateHistory(ctx, history)

	// Log Activity
	u.logActivity(ctx, userID, "Assign Inventory Item", fmt.Sprintf("Assigned item ID: %d (SN: %s) to pelanggan: %s (ID: %d)", item.ID, item.SerialNumber, pelanggan.Nama, pelanggan.ID))

	return nil
}

func (u *inventoryUsecase) UnassignItem(ctx context.Context, itemID uint64, notes string, userID uint64) error {
	item, err := u.repo.GetItemByID(ctx, itemID)
	if err != nil {
		return err
	}

	if item.PelangganID == nil {
		return errors.New("item is not assigned to any pelanggan")
	}

	oldPelangganID := *item.PelangganID
	var oldPelangganName string = "Unknown"
	if item.Pelanggan != nil {
		oldPelangganName = item.Pelanggan.Nama
	}

	// Get available status ID
	availableStatus, err := u.repo.GetStatusByName(ctx, "available")
	if err != nil || availableStatus == nil {
		return errors.New("available status configuration not found in database")
	}

	// Update item
	item.StatusID = availableStatus.ID
	item.PelangganID = nil
	if notes != "" {
		if item.Notes != nil && *item.Notes != "" {
			newNotes := fmt.Sprintf("%s\nUnassigned note: %s", *item.Notes, notes)
			item.Notes = &newNotes
		} else {
			item.Notes = &notes
		}
	}

	err = u.repo.UpdateItem(ctx, item)
	if err != nil {
		return err
	}

	// Log History
	actionText := fmt.Sprintf("Unassigned from pelanggan %s (ID: %d)", oldPelangganName, oldPelangganID)
	history := &domain.InventoryHistory{
		ItemID:    item.ID,
		UserID:    userID,
		Action:    actionText,
		Timestamp: time.Now(),
	}
	_ = u.repo.CreateHistory(ctx, history)

	// Log Activity
	u.logActivity(ctx, userID, "Unassign Inventory Item", fmt.Sprintf("Unassigned item ID: %d (SN: %s) from pelanggan ID: %d", item.ID, item.SerialNumber, oldPelangganID))

	return nil
}

func (u *inventoryUsecase) FetchHistoryByItemID(ctx context.Context, itemID uint64) ([]domain.InventoryHistory, error) {
	return u.repo.GetHistoryByItemID(ctx, itemID)
}

func (u *inventoryUsecase) CreateItemType(ctx context.Context, itemType *domain.InventoryItemType) (*domain.InventoryItemType, error) {
	if itemType.Name == "" {
		return nil, errors.New("name is required")
	}
	existing, err := u.repo.GetItemTypeByName(ctx, itemType.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("item type name already exists")
	}
	err = u.repo.CreateItemType(ctx, itemType)
	if err != nil {
		return nil, err
	}
	return itemType, nil
}

func (u *inventoryUsecase) UpdateItemType(ctx context.Context, id uint64, name string) (*domain.InventoryItemType, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	existing, err := u.repo.GetItemTypeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	existingByName, err := u.repo.GetItemTypeByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existingByName != nil && existingByName.ID != id {
		return nil, errors.New("item type name already exists")
	}
	existing.Name = name
	err = u.repo.UpdateItemType(ctx, existing)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (u *inventoryUsecase) DeleteItemType(ctx context.Context, id uint64) error {
	_, err := u.repo.GetItemTypeByID(ctx, id)
	if err != nil {
		return err
	}
	return u.repo.DeleteItemType(ctx, id)
}

func (u *inventoryUsecase) CreateStatus(ctx context.Context, status *domain.InventoryStatus) (*domain.InventoryStatus, error) {
	if status.Name == "" {
		return nil, errors.New("name is required")
	}
	existing, err := u.repo.GetStatusByName(ctx, status.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("status name already exists")
	}
	err = u.repo.CreateStatus(ctx, status)
	if err != nil {
		return nil, err
	}
	return status, nil
}

func (u *inventoryUsecase) UpdateStatus(ctx context.Context, id uint64, name string) (*domain.InventoryStatus, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	existing, err := u.repo.GetStatusByID(ctx, id)
	if err != nil {
		return nil, err
	}
	existingByName, err := u.repo.GetStatusByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existingByName != nil && existingByName.ID != id {
		return nil, errors.New("status name already exists")
	}
	existing.Name = name
	err = u.repo.UpdateStatus(ctx, existing)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (u *inventoryUsecase) DeleteStatus(ctx context.Context, id uint64) error {
	_, err := u.repo.GetStatusByID(ctx, id)
	if err != nil {
		return err
	}
	return u.repo.DeleteStatus(ctx, id)
}

func (u *inventoryUsecase) FetchGlobalHistory(ctx context.Context) ([]domain.InventoryHistory, error) {
	return u.repo.GetGlobalHistory(ctx)
}


func (u *inventoryUsecase) BulkImport(ctx context.Context, items []domain.InventoryItem, userID uint64) (int, int, []string, error) {
	successCount := 0
	errorCount := 0
	var errorsList []string

	// Fetch all existing SNs and MACs at once to prevent N+1 queries inside loop
	existingSNs, existingMacs, err := u.repo.GetAllSerialNumbersAndMacs(ctx)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("gagal memuat data inventaris yang ada: %w", err)
	}

	for i, item := range items {
		normSN := strings.ToUpper(item.SerialNumber)
		// Verify serial number uniqueness in-memory
		if existingSNs[normSN] {
			errorsList = append(errorsList, fmt.Sprintf("Baris %d: Serial number %s sudah ada", i+2, item.SerialNumber))
			errorCount++
			continue
		}

		// Verify MAC address uniqueness in-memory
		if item.MacAddress != nil && *item.MacAddress != "" {
			normMac := strings.ToUpper(*item.MacAddress)
			if existingMacs[normMac] {
				errorsList = append(errorsList, fmt.Sprintf("Baris %d: MAC address %s sudah ada", i+2, *item.MacAddress))
				errorCount++
				continue
			}
		}

		// Create
		err = u.repo.CreateItem(ctx, &item)
		if err != nil {
			errorsList = append(errorsList, fmt.Sprintf("Baris %d: Gagal menyimpan item: %v", i+2, err))
			errorCount++
			continue
		}

		// Update in-memory cache to prevent duplicate rows in the imported file from passing uniqueness check
		existingSNs[normSN] = true
		if item.MacAddress != nil && *item.MacAddress != "" {
			existingMacs[strings.ToUpper(*item.MacAddress)] = true
		}

		// Log History
		history := &domain.InventoryHistory{
			ItemID:    item.ID,
			UserID:    userID,
			Action:    "Imported item",
			Timestamp: time.Now(),
		}
		_ = u.repo.CreateHistory(ctx, history)

		// Log Activity
		u.logActivity(ctx, userID, "Import Inventory Item", fmt.Sprintf("Imported item with SN: %s", item.SerialNumber))

		successCount++
	}

	return successCount, errorCount, errorsList, nil
}
