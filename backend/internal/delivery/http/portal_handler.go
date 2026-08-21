package http

import (
	"net/http"
	"strings"

	"billing-backend/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PortalHandler struct {
	db *gorm.DB
}

func NewPortalHandler(db *gorm.DB) *PortalHandler {
	return &PortalHandler{db: db}
}

// CustomerLookup handles ultra-fast single-query customer data lookup for Portal Pelanggan.
// Supports lookup by Email or Phone Number (handling +62, 62, 08 formats).
func (h *PortalHandler) CustomerLookup(c *gin.Context) {
	identifier := strings.TrimSpace(c.Query("identifier"))
	if identifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'identifier' (email or phone) is required"})
		return
	}

	dbVal, exists := c.Get("db")
	db := h.db
	if exists {
		if gormDB, ok := dbVal.(*gorm.DB); ok {
			db = gormDB
		}
	}

	var customer domain.Pelanggan
	var err error

	isEmail := strings.Contains(identifier, "@")

	if isEmail {
		err = db.WithContext(c.Request.Context()).
			Preload("Langganan.PaketLayanan").
			Preload("Invoices", func(db *gorm.DB) *gorm.DB {
				return db.Order("tgl_invoice DESC, id DESC").Limit(100)
			}).
			Where("LOWER(email) = ?", strings.ToLower(identifier)).
			First(&customer).Error
	} else {
		// Phone normalization: Generate candidates (08xxx, 628xxx, +628xxx, 8xxx)
		cleaned := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(identifier, "-", ""), " ", ""), "+", "")
		var candidates []string

		if strings.HasPrefix(cleaned, "628") {
			core := cleaned[2:] // 8xxx
			candidates = append(candidates, "0"+core, "62"+core, "+62"+core, core)
		} else if strings.HasPrefix(cleaned, "08") {
			core := cleaned[1:] // 8xxx
			candidates = append(candidates, "0"+core, "62"+core, "+62"+core, core)
		} else if strings.HasPrefix(cleaned, "8") {
			candidates = append(candidates, "0"+cleaned, "62"+cleaned, "+62"+cleaned, cleaned)
		} else {
			candidates = append(candidates, identifier, cleaned)
		}

		err = db.WithContext(c.Request.Context()).
			Preload("Langganan.PaketLayanan").
			Preload("Invoices", func(db *gorm.DB) *gorm.DB {
				return db.Order("tgl_invoice DESC, id DESC").Limit(100)
			}).
			Where("no_telp IN ?", candidates).
			First(&customer).Error
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found with provided identifier"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query customer data: " + err.Error()})
		return
	}

	// Select primary active subscription if available, or first subscription
	var primaryLangganan *domain.Langganan
	if len(customer.Langganan) > 0 {
		for i := range customer.Langganan {
			if strings.EqualFold(customer.Langganan[i].Status, "Aktif") {
				primaryLangganan = &customer.Langganan[i]
				break
			}
		}
		if primaryLangganan == nil {
			primaryLangganan = &customer.Langganan[0]
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"pelanggan": customer,
		"langganan": primaryLangganan,
		"invoices":  customer.Invoices,
	}})
}
