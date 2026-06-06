package http

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"billing-backend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type InventoryHandler struct {
	usecase domain.InventoryUsecase
}

func NewInventoryHandler(r *gin.RouterGroup, iu domain.InventoryUsecase, authMiddleware gin.HandlerFunc) {
	handler := &InventoryHandler{
		usecase: iu,
	}

	// Inventory Items routes
	inventoryGroup := r.Group("/inventory")
	inventoryGroup.Use(authMiddleware)
	{
		inventoryGroup.GET("", handler.GetItems)
		inventoryGroup.GET("/", handler.GetItems)
		inventoryGroup.GET("/:id", handler.GetItemByID)
		inventoryGroup.POST("", handler.CreateItem)
		inventoryGroup.POST("/", handler.CreateItem)
		inventoryGroup.PATCH("/:id", handler.UpdateItem)
		inventoryGroup.PUT("/:id", handler.UpdateItem)
		inventoryGroup.DELETE("/:id", handler.DeleteItem)

		// Assignment & History
		inventoryGroup.GET("/history/all", handler.GetGlobalHistory)
		inventoryGroup.GET("/:id/history", handler.GetItemHistory)
		inventoryGroup.POST("/:id/assign", handler.AssignItem)
		inventoryGroup.POST("/:id/unassign", handler.UnassignItem)

		// Bulk Import
		inventoryGroup.POST("/bulk-import", handler.BulkImport)
	}

	// Inventory Types routes
	typesGroup := r.Group("/inventory-types")
	typesGroup.Use(authMiddleware)
	{
		typesGroup.GET("", handler.GetItemTypes)
		typesGroup.GET("/", handler.GetItemTypes)
		typesGroup.POST("", handler.CreateItemType)
		typesGroup.POST("/", handler.CreateItemType)
		typesGroup.PATCH("/:id", handler.UpdateItemType)
		typesGroup.DELETE("/:id", handler.DeleteItemType)
	}

	// Inventory Statuses routes
	statusesGroup := r.Group("/inventory-statuses")
	statusesGroup.Use(authMiddleware)
	{
		statusesGroup.GET("", handler.GetStatuses)
		statusesGroup.GET("/", handler.GetStatuses)
		statusesGroup.POST("", handler.CreateStatus)
		statusesGroup.POST("/", handler.CreateStatus)
		statusesGroup.PATCH("/:id", handler.UpdateStatus)
		statusesGroup.DELETE("/:id", handler.DeleteStatus)
	}
}

func (h *InventoryHandler) GetItems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	var itemTypeIDPtr, statusIDPtr, pelangganIDPtr *uint64

	if typeStr := c.Query("item_type_id"); typeStr != "" {
		if val, err := strconv.ParseUint(typeStr, 10, 64); err == nil {
			itemTypeIDPtr = &val
		}
	}

	if statusStr := c.Query("status_id"); statusStr != "" {
		if val, err := strconv.ParseUint(statusStr, 10, 64); err == nil {
			statusIDPtr = &val
		}
	}

	if pelangganStr := c.Query("pelanggan_id"); pelangganStr != "" {
		if val, err := strconv.ParseUint(pelangganStr, 10, 64); err == nil {
			pelangganIDPtr = &val
		}
	}

	items, total, err := h.usecase.FetchItems(c.Request.Context(), page, pageSize, search, itemTypeIDPtr, statusIDPtr, pelangganIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *InventoryHandler) GetItemByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	item, err := h.usecase.GetItemByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *InventoryHandler) CreateItem(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	var item domain.InventoryItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	newItem, err := h.usecase.CreateItem(c.Request.Context(), &item, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": newItem})
}

func (h *InventoryHandler) UpdateItem(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var item domain.InventoryItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	updatedItem, err := h.usecase.UpdateItem(c.Request.Context(), id, &item, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updatedItem})
}

func (h *InventoryHandler) DeleteItem(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.usecase.DeleteItem(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory item deleted successfully"})
}

func (h *InventoryHandler) GetItemTypes(c *gin.Context) {
	types, err := h.usecase.FetchItemTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": types})
}

func (h *InventoryHandler) GetStatuses(c *gin.Context) {
	statuses, err := h.usecase.FetchStatuses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": statuses})
}

func (h *InventoryHandler) GetItemHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	history, err := h.usecase.FetchHistoryByItemID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": history})
}

type AssignItemRequest struct {
	PelangganID uint64 `json:"pelanggan_id" binding:"required"`
	Notes       string `json:"notes"`
}

func (h *InventoryHandler) AssignItem(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req AssignItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	err = h.usecase.AssignItem(c.Request.Context(), id, req.PelangganID, req.Notes, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory item assigned successfully"})
}

type UnassignItemRequest struct {
	Notes string `json:"notes"`
}

func (h *InventoryHandler) UnassignItem(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req UnassignItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	err = h.usecase.UnassignItem(c.Request.Context(), id, req.Notes, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory item unassigned successfully"})
}

func (h *InventoryHandler) CreateItemType(c *gin.Context) {
	var itemType domain.InventoryItemType
	if err := c.ShouldBindJSON(&itemType); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	created, err := h.usecase.CreateItemType(c.Request.Context(), &itemType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": created})
}

func (h *InventoryHandler) UpdateItemType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.usecase.UpdateItemType(c.Request.Context(), id, input.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updated})
}

func (h *InventoryHandler) DeleteItemType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.usecase.DeleteItemType(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item type deleted successfully"})
}

func (h *InventoryHandler) CreateStatus(c *gin.Context) {
	var status domain.InventoryStatus
	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	created, err := h.usecase.CreateStatus(c.Request.Context(), &status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": created})
}

func (h *InventoryHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.usecase.UpdateStatus(c.Request.Context(), id, input.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updated})
}

func (h *InventoryHandler) DeleteStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.usecase.DeleteStatus(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status deleted successfully"})
}

func (h *InventoryHandler) GetGlobalHistory(c *gin.Context) {
	history, err := h.usecase.FetchGlobalHistory(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format response to match frontend expectation
	type GlobalHistoryResponse struct {
		ID           uint64    `json:"id"`
		ItemID       uint64    `json:"item_id"`
		Action       string    `json:"action"`
		Timestamp    time.Time `json:"timestamp"`
		SerialNumber string    `json:"serial_number"`
		MacAddress   string    `json:"mac_address"`
		User         struct {
			ID   uint64 `json:"id"`
			Name string `json:"name"`
		} `json:"user"`
	}

	resp := make([]GlobalHistoryResponse, 0, len(history))
	for _, hist := range history {
		var sn, mac string
		if hist.InventoryItem != nil {
			sn = hist.InventoryItem.SerialNumber
			if hist.InventoryItem.MacAddress != nil {
				mac = *hist.InventoryItem.MacAddress
			}
		}
		item := GlobalHistoryResponse{
			ID:           hist.ID,
			ItemID:       hist.ItemID,
			Action:       hist.Action,
			Timestamp:    hist.Timestamp,
			SerialNumber: sn,
			MacAddress:   mac,
		}
		if hist.User != nil {
			item.User.ID = hist.User.ID
			item.User.Name = hist.User.Name
		} else {
			item.User.Name = "System"
		}
		resp = append(resp, item)
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func findColumnIndex(headers []string, possibleNames []string) int {
	for i, h := range headers {
		normH := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(h)), " ", "_"), "-", "_"), ".", "_")
		for _, name := range possibleNames {
			normName := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(name)), " ", "_"), "-", "_"), ".", "_")
			if normH == normName {
				return i
			}
		}
	}
	return -1
}

func cleanMacAddress(mac string) string {
	cleaned := ""
	for _, char := range mac {
		if (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F') {
			cleaned += string(char)
		}
	}
	cleaned = strings.ToUpper(cleaned)
	if len(cleaned) == 12 {
		parts := make([]string, 6)
		for i := 0; i < 6; i++ {
			parts[i] = cleaned[i*2 : i*2+2]
		}
		return strings.Join(parts, ":")
	}
	return mac
}

func (h *InventoryHandler) BulkImport(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := strconv.ParseUint(userIDStr.(string), 10, 64)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".xlsx" && ext != ".xls" && ext != ".csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File harus berformat .xlsx, .xls, atau .csv"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membuka file: " + err.Error()})
		return
	}
	defer openedFile.Close()

	var rows [][]string

	if ext == ".csv" {
		reader := csv.NewReader(openedFile)
		reader.LazyQuotes = true
		reader.FieldsPerRecord = -1
		rows, err = reader.ReadAll()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading CSV: " + err.Error()})
			return
		}
	} else {
		// Excel
		xlsx, err := excelize.OpenReader(openedFile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading Excel: " + err.Error()})
			return
		}
		defer xlsx.Close()

		sheets := xlsx.GetSheetList()
		if len(sheets) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Excel does not contain any sheets"})
			return
		}
		rows, err = xlsx.GetRows(sheets[0])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error getting Excel rows: " + err.Error()})
			return
		}
	}

	if len(rows) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File data kosong atau tidak memiliki baris data"})
		return
	}

	headers := rows[0]

	serialIdx := findColumnIndex(headers, []string{
		"serial_number", "serialnumber", "serial_no", "serial no", "no serial", "nomor serial", "sn", "serial number",
		"Serial_Number", "SerialNumber", "Serial_no", "Serial no", "No Serial", "Nomor Serial", "Serial Number",
	})
	if serialIdx == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Kolom wajib tidak ditemukan: Serial Number. Kolom yang ditemukan: %v", headers)})
		return
	}

	itemTypeIdx := findColumnIndex(headers, []string{
		"item_type", "itemtype", "item_type_id", "itemtypeid", "jenis_barang", "tipe_barang", "type", "tipe",
		"Item_Type", "ItemType", "Item_type", "Itemtype", "Jenis_Barang", "Tipe_Barang", "Type", "Tipe",
		"tipe barang", "tipe_barang",
	})
	if itemTypeIdx == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Kolom wajib tidak ditemukan: Tipe Barang/Item Type. Kolom yang ditemukan: %v", headers)})
		return
	}

	statusIdx := findColumnIndex(headers, []string{
		"status", "status_id", "statusid", "kondisi", "keadaan", "Status", "Status_id", "Kondisi", "Keadaan", "status id",
	})
	if statusIdx == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Kolom wajib tidak ditemukan: Status. Kolom yang ditemukan: %v", headers)})
		return
	}

	macIdx := findColumnIndex(headers, []string{"mac_address", "mac", "macaddress", "alamat_mac", "mac addr"})
	locationIdx := findColumnIndex(headers, []string{"location", "lokasi", "tempat", "letak"})
	notesIdx := findColumnIndex(headers, []string{"notes", "catatan", "keterangan", "note"})
	purchaseDateIdx := findColumnIndex(headers, []string{"purchase_date", "tanggal_pembelian", "tgl_pembelian", "purchasedate"})

	types, err := h.usecase.FetchItemTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch item types: " + err.Error()})
		return
	}
	validItemTypes := make(map[string]uint64)
	validItemTypeIDs := make(map[uint64]bool)
	for _, t := range types {
		validItemTypes[strings.ToLower(t.Name)] = t.ID
		validItemTypeIDs[t.ID] = true
	}

	statuses, err := h.usecase.FetchStatuses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statuses: " + err.Error()})
		return
	}
	validStatuses := make(map[string]uint64)
	validStatusIDs := make(map[uint64]bool)
	for _, s := range statuses {
		validStatuses[strings.ToLower(s.Name)] = s.ID
		validStatusIDs[s.ID] = true
	}

	var defaultTypeID uint64
	for _, t := range types {
		if strings.Contains(strings.ToLower(t.Name), "ont") && strings.Contains(strings.ToLower(t.Name), "zte") {
			defaultTypeID = t.ID
			break
		}
	}
	if defaultTypeID == 0 && len(types) > 0 {
		defaultTypeID = types[0].ID
	}

	var defaultStatusID uint64
	for _, s := range statuses {
		if strings.Contains(strings.ToLower(s.Name), "gudang") {
			defaultStatusID = s.ID
			break
		}
	}
	if defaultStatusID == 0 && len(statuses) > 0 {
		defaultStatusID = statuses[0].ID
	}

	resolveType := func(val string) uint64 {
		val = strings.TrimSpace(val)
		if val == "" {
			return defaultTypeID
		}
		if id, err := strconv.ParseUint(val, 10, 64); err == nil {
			if validItemTypeIDs[id] {
				return id
			}
		}
		if id, exists := validItemTypes[strings.ToLower(val)]; exists {
			return id
		}
		return defaultTypeID
	}

	resolveStatus := func(val string) uint64 {
		val = strings.TrimSpace(val)
		if val == "" {
			return defaultStatusID
		}
		if id, err := strconv.ParseUint(val, 10, 64); err == nil {
			if validStatusIDs[id] {
				return id
			}
		}
		if id, exists := validStatuses[strings.ToLower(val)]; exists {
			return id
		}
		return defaultStatusID
	}

	var itemsToImport []domain.InventoryItem
	var preValidationErrors []string

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}

		var serialNumber string
		if serialIdx < len(row) {
			serialNumber = strings.ToUpper(strings.TrimSpace(row[serialIdx]))
		}
		if serialNumber == "" {
			continue
		}

		var macAddress *string
		if macIdx >= 0 && macIdx < len(row) {
			mVal := strings.TrimSpace(row[macIdx])
			if mVal != "" {
				cleaned := cleanMacAddress(mVal)
				macAddress = &cleaned
			}
		}

		var location *string
		if locationIdx >= 0 && locationIdx < len(row) {
			locVal := strings.TrimSpace(row[locationIdx])
			if locVal != "" {
				location = &locVal
			}
		}

		var notes *string
		if notesIdx >= 0 && notesIdx < len(row) {
			notesVal := strings.TrimSpace(row[notesIdx])
			if notesVal != "" {
				notes = &notesVal
			}
		}

		var purchaseDate *time.Time
		if purchaseDateIdx >= 0 && purchaseDateIdx < len(row) {
			pDateVal := strings.TrimSpace(row[purchaseDateIdx])
			if pDateVal != "" {
				var parsedDate time.Time
				var pErr error
				formats := []string{"2006-01-02", "02/01/2006", "02-01-2006"}
				for _, f := range formats {
					parsedDate, pErr = time.Parse(f, pDateVal)
					if pErr == nil {
						break
					}
				}
				if pErr != nil {
					preValidationErrors = append(preValidationErrors, fmt.Sprintf("Baris %d: Format tanggal pembelian tidak valid (gunakan YYYY-MM-DD)", i+1))
					continue
				}
				purchaseDate = &parsedDate
			}
		}

		var itemTypeID uint64
		if itemTypeIdx < len(row) {
			itemTypeID = resolveType(row[itemTypeIdx])
		} else {
			itemTypeID = defaultTypeID
		}

		var statusID uint64
		if statusIdx < len(row) {
			statusID = resolveStatus(row[statusIdx])
		} else {
			statusID = defaultStatusID
		}

		item := domain.InventoryItem{
			SerialNumber: serialNumber,
			MacAddress:   macAddress,
			Location:     location,
			Notes:        notes,
			PurchaseDate: purchaseDate,
			ItemTypeID:   itemTypeID,
			StatusID:     statusID,
		}
		itemsToImport = append(itemsToImport, item)
	}

	if len(preValidationErrors) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"success":       false,
			"message":       "Gagal mengimport file karena kesalahan validasi format data",
			"success_count": 0,
			"error_count":   len(preValidationErrors),
			"errors":        preValidationErrors,
		})
		return
	}

	successCount, errorCount, errorsList, err := h.usecase.BulkImport(c.Request.Context(), itemsToImport, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	message := fmt.Sprintf("Import selesai! %d item berhasil ditambahkan, %d item gagal.", successCount, errorCount)
	c.JSON(http.StatusOK, gin.H{
		"success":       errorCount == 0,
		"message":       message,
		"success_count": successCount,
		"error_count":   errorCount,
		"errors":        errorsList,
	})
}
