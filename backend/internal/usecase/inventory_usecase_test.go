package usecase

import (
	"context"
	"errors"
	"testing"

	"billing-backend/internal/domain"
)

type mockInventoryRepo struct {
	domain.InventoryRepository
	items        map[uint64]*domain.InventoryItem
	types        map[uint64]*domain.InventoryItemType
	statuses     map[uint64]*domain.InventoryStatus
	history      []*domain.InventoryHistory
	serialNumber map[string]*domain.InventoryItem
	macAddress   map[string]*domain.InventoryItem
}

func (m *mockInventoryRepo) GetItems(ctx context.Context, limit, offset int, search string, itemTypeID, statusID, pelangganID *uint64) ([]domain.InventoryItem, int64, error) {
	var result []domain.InventoryItem
	for _, item := range m.items {
		if itemTypeID != nil && item.ItemTypeID != *itemTypeID {
			continue
		}
		if statusID != nil && item.StatusID != *statusID {
			continue
		}
		if pelangganID != nil {
			if item.PelangganID == nil || *item.PelangganID != *pelangganID {
				continue
			}
		}
		result = append(result, *item)
	}
	return result, int64(len(result)), nil
}

func (m *mockInventoryRepo) GetItemByID(ctx context.Context, id uint64) (*domain.InventoryItem, error) {
	item, exists := m.items[id]
	if !exists {
		return nil, errors.New("not found")
	}
	if item.StatusID != 0 {
		item.Status = m.statuses[item.StatusID]
	}
	if item.ItemTypeID != 0 {
		item.ItemType = m.types[item.ItemTypeID]
	}
	return item, nil
}

func (m *mockInventoryRepo) GetItemBySerialNumber(ctx context.Context, sn string) (*domain.InventoryItem, error) {
	item, exists := m.serialNumber[sn]
	if !exists {
		return nil, nil
	}
	if item.StatusID != 0 {
		item.Status = m.statuses[item.StatusID]
	}
	if item.ItemTypeID != 0 {
		item.ItemType = m.types[item.ItemTypeID]
	}
	return item, nil
}

func (m *mockInventoryRepo) GetItemByMacAddress(ctx context.Context, mac string) (*domain.InventoryItem, error) {
	item, exists := m.macAddress[mac]
	if !exists {
		return nil, nil
	}
	if item.StatusID != 0 {
		item.Status = m.statuses[item.StatusID]
	}
	if item.ItemTypeID != 0 {
		item.ItemType = m.types[item.ItemTypeID]
	}
	return item, nil
}

func (m *mockInventoryRepo) CreateItem(ctx context.Context, item *domain.InventoryItem) error {
	m.items[item.ID] = item
	m.serialNumber[item.SerialNumber] = item
	if item.MacAddress != nil {
		m.macAddress[*item.MacAddress] = item
	}
	return nil
}

func (m *mockInventoryRepo) UpdateItem(ctx context.Context, item *domain.InventoryItem) error {
	m.items[item.ID] = item
	m.serialNumber[item.SerialNumber] = item
	if item.MacAddress != nil {
		m.macAddress[*item.MacAddress] = item
	}
	return nil
}

func (m *mockInventoryRepo) GetItemTypeByID(ctx context.Context, id uint64) (*domain.InventoryItemType, error) {
	t, exists := m.types[id]
	if !exists {
		return nil, errors.New("type not found")
	}
	return t, nil
}

func (m *mockInventoryRepo) GetStatusByID(ctx context.Context, id uint64) (*domain.InventoryStatus, error) {
	s, exists := m.statuses[id]
	if !exists {
		return nil, errors.New("status not found")
	}
	return s, nil
}

func (m *mockInventoryRepo) GetStatusByName(ctx context.Context, name string) (*domain.InventoryStatus, error) {
	for _, s := range m.statuses {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, nil
}

func (m *mockInventoryRepo) CreateHistory(ctx context.Context, h *domain.InventoryHistory) error {
	m.history = append(m.history, h)
	return nil
}

func (m *mockInventoryRepo) GetHistoryByItemID(ctx context.Context, itemID uint64) ([]domain.InventoryHistory, error) {
	var res []domain.InventoryHistory
	for _, h := range m.history {
		if h.ItemID == itemID {
			res = append(res, *h)
		}
	}
	return res, nil
}

type mockPelangganRepositoryForInventory struct {
	domain.PelangganRepository
	pelanggans map[uint64]*domain.Pelanggan
}

func (m *mockPelangganRepositoryForInventory) GetByID(ctx context.Context, id uint64) (*domain.Pelanggan, error) {
	p, exists := m.pelanggans[id]
	if !exists {
		return nil, errors.New("pelanggan not found")
	}
	return p, nil
}

func TestInventoryUsecase_AssignAndUnassign(t *testing.T) {
	// Setup mocks
	itemTypes := map[uint64]*domain.InventoryItemType{
		1: {ID: 1, Name: "Router"},
	}
	statuses := map[uint64]*domain.InventoryStatus{
		1: {ID: 1, Name: "available"},
		2: {ID: 2, Name: "assigned"},
	}
	items := map[uint64]*domain.InventoryItem{
		10: {
			ID:           10,
			SerialNumber: "SN123456",
			ItemTypeID:   1,
			StatusID:     1, // available
			Status:       statuses[1],
		},
	}
	pelanggans := map[uint64]*domain.Pelanggan{
		100: {
			ID:   100,
			Nama: "John Doe",
		},
	}

	repo := &mockInventoryRepo{
		items:        items,
		types:        itemTypes,
		statuses:     statuses,
		serialNumber: make(map[string]*domain.InventoryItem),
		macAddress:   make(map[string]*domain.InventoryItem),
	}
	// populate SN lookup
	for _, item := range items {
		repo.serialNumber[item.SerialNumber] = item
	}

	pelangganRepo := &mockPelangganRepositoryForInventory{
		pelanggans: pelanggans,
	}

	systemRepo := &mockSystemRepo{}

	uc := NewInventoryUsecase(repo, pelangganRepo, systemRepo)

	ctx := context.Background()

	// 1. Assign available item to pelanggan 100
	err := uc.AssignItem(ctx, 10, 100, "For living room", 1)
	if err != nil {
		t.Fatalf("expected no error during AssignItem, got %v", err)
	}

	// Verify status updated to assigned (2) and pelanggan id set
	assignedItem, _ := repo.GetItemByID(ctx, 10)
	if assignedItem.StatusID != 2 {
		t.Errorf("expected StatusID to be 2, got %d", assignedItem.StatusID)
	}
	if assignedItem.PelangganID == nil || *assignedItem.PelangganID != 100 {
		t.Errorf("expected PelangganID to be 100, got %v", assignedItem.PelangganID)
	}

	// Verify history added
	histories, _ := repo.GetHistoryByItemID(ctx, 10)
	if len(histories) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(histories))
	} else if histories[0].Action != "Assigned to pelanggan John Doe (ID: 100)" {
		t.Errorf("unexpected history action: %s", histories[0].Action)
	}

	// 2. Try to assign the already assigned item again (should fail)
	err = uc.AssignItem(ctx, 10, 100, "Try again", 1)
	if err == nil {
		t.Fatal("expected error when assigning already assigned item, got nil")
	}

	// 3. Unassign item
	err = uc.UnassignItem(ctx, 10, "Returned by customer", 1)
	if err != nil {
		t.Fatalf("expected no error during UnassignItem, got %v", err)
	}

	// Verify status updated back to available (1) and pelanggan id nil
	unassignedItem, _ := repo.GetItemByID(ctx, 10)
	if unassignedItem.StatusID != 1 {
		t.Errorf("expected StatusID to be 1, got %d", unassignedItem.StatusID)
	}
	if unassignedItem.PelangganID != nil {
		t.Errorf("expected PelangganID to be nil, got %v", unassignedItem.PelangganID)
	}

	// Verify another history added
	histories, _ = repo.GetHistoryByItemID(ctx, 10)
	if len(histories) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(histories))
	}
}
