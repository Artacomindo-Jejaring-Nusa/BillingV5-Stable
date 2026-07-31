package database

import (
	"log"
	"strings"
	"time"

	"billing-backend/internal/domain"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed populates the database with default roles, permissions, and a super admin user.
func Seed(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. Generate Permission List
		var permissions []string

		// --- Actions for Menus ---
		actions := []string{"create", "view", "edit", "delete"}
		menus := []string{
			"Dashboard",
			"Pelanggan",
			"Langganan",
			"Data_Teknis",
			"Brand_&_Paket",
			"Invoices",
			"Reports_Revenue",
			"Mikrotik_Servers",
			"Users",
			"Roles",
			"Permissions",
			"SK",
			"Simulasi_Harga",
			"Inventory",
			"Dashboard_Pelanggan",
			"Activity_Log",
			"OLT",
			"ODP_Management",
			"Trouble_Tickets",
			"Diskon",
			"AI_Analytics",
		}
		for _, menu := range menus {
			cleaned := strings.ToLower(menu)
			cleaned = strings.ReplaceAll(cleaned, " & ", "_")
			cleaned = strings.ReplaceAll(cleaned, " ", "_")
			for _, action := range actions {
				permissions = append(permissions, action+"_"+cleaned)
			}
		}

		// --- Widgets ---
		widgets := []string{
			"pendapatan_bulanan",
			"statistik_pelanggan",
			"statistik_server",
			"pelanggan_per_lokasi",
			"pelanggan_per_paket",
			"tren_pertumbuhan",
			"invoice_bulanan",
			"status_langganan",
			"alamat_aktif",
			"invoice_generation_monitor",
			"future_invoice_projection",
			"pelanggan_statistik_utama",
			"pelanggan_pendapatan_jakinet",
			"pelanggan_distribusi_chart",
			"pelanggan_pertumbuhan_chart",
			"pelanggan_status_overview_chart",
			"pelanggan_metrik_cepat",
			"pelanggan_tren_pendapatan_chart",
		}
		for _, widget := range widgets {
			permissions = append(permissions, "view_widget_"+widget)
		}

		// --- System Features ---
		features := []string{
			"settings",
			"uploads",
			"traffic_monitoring",
		}
		for _, feature := range features {
			for _, action := range actions {
				permissions = append(permissions, action+"_"+feature)
			}
		}

		// --- Special Management Actions ---
		permissions = append(permissions, "manage_sk")
		permissions = append(permissions, "manage_settings")

		// 2. Insert Permissions into DB
		var dbPermissions []domain.Permission
		for _, pName := range permissions {
			var perm domain.Permission
			err := tx.Where("name = ?", pName).First(&perm).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					perm = domain.Permission{Name: pName}
					if err := tx.Create(&perm).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}
			dbPermissions = append(dbPermissions, perm)
		}

		// 3. Create Superadmin Role
		var superadminRole domain.Role
		err := tx.Preload("Permissions").Where("name = ?", "superadmin").First(&superadminRole).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				superadminRole = domain.Role{
					Name:        "superadmin",
					Permissions: dbPermissions,
				}
				if err := tx.Create(&superadminRole).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// Update permissions if role already exists to ensure it has all
			if err := tx.Model(&superadminRole).Association("Permissions").Replace(dbPermissions); err != nil {
				return err
			}
		}

		// 4. Create or Update Super Admin User
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		now := time.Now()

		var superadminUser domain.User
		err = tx.Where("email = ?", "admin@example.com").First(&superadminUser).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				superadminUser = domain.User{
					Name:              "Super Admin",
					Email:             "admin@example.com",
					Password:          string(hashedPassword),
					IsActive:          true,
					PasswordChangedAt: &now,
					RoleID:            &superadminRole.ID,
					CreatedAt:         &now,
					UpdatedAt:         &now,
				}
				if err := tx.Create(&superadminUser).Error; err != nil {
					return err
				}
				log.Println("✅ Super Admin user seeded successfully! Email: admin@example.com / Password: password")
			} else {
				return err
			}
		} else {
			// Ensure password is updated to 'password'
			tx.Model(&superadminUser).Updates(map[string]interface{}{
				"password":  string(hashedPassword),
				"is_active": true,
				"role_id":   superadminRole.ID,
			})
			log.Println("✅ Super Admin password reset to 'password'!")
		}

		// 5. Seed Inventory Statuses
		statuses := []string{"available", "assigned", "broken", "maintenance"}
		for _, sName := range statuses {
			var status domain.InventoryStatus
			err := tx.Where("name = ?", sName).First(&status).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					status = domain.InventoryStatus{Name: sName}
					if err := tx.Create(&status).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}

		// 6. Seed Inventory Item Types
		itemTypes := []string{"Router", "ONT/ONU", "STB", "Switch", "Access Point"}
		for _, tName := range itemTypes {
			var itemType domain.InventoryItemType
			err := tx.Where("name = ?", tName).First(&itemType).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					itemType = domain.InventoryItemType{Name: tName}
					if err := tx.Create(&itemType).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}

		// 7. Seed Dummy Brands (HargaLayanan)
		brands := []domain.HargaLayanan{
			{IDBrand: "JAKINET", Brand: "JAKINET", Pajak: 11.0, XenditKeyName: "JAKINET"},
			{IDBrand: "JELANTIK", Brand: "JELANTIK", Pajak: 11.0, XenditKeyName: "JELANTIK"},
		}
		for _, b := range brands {
			var existing domain.HargaLayanan
			if err := tx.Where("id_brand = ?", b.IDBrand).First(&existing).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					tx.Create(&b)
				}
			}
		}

		// 8. Seed Dummy Paket Layanan
		pakets := []domain.PaketLayanan{
			{IDBrand: "JAKINET", NamaPaket: "Jakinet Lite 10Mbps", Kecepatan: 10, Harga: 150000},
			{IDBrand: "JAKINET", NamaPaket: "Jakinet Pro 20Mbps", Kecepatan: 20, Harga: 250000},
			{IDBrand: "JELANTIK", NamaPaket: "Jelantik Express 15Mbps", Kecepatan: 15, Harga: 180000},
			{IDBrand: "JELANTIK", NamaPaket: "Jelantik Ultima 30Mbps", Kecepatan: 30, Harga: 300000},
		}
		for _, pk := range pakets {
			var existing domain.PaketLayanan
			if err := tx.Where("id_brand = ? AND nama_paket = ?", pk.IDBrand, pk.NamaPaket).First(&existing).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					tx.Create(&pk)
				}
			}
		}

		// 9. Seed Dummy Pelanggan & DataTeknis
		var countPelanggan int64
		tx.Model(&domain.Pelanggan{}).Count(&countPelanggan)
		if countPelanggan == 0 {
			brandJakinet := "JAKINET"
			dummyPelanggan := domain.Pelanggan{
				NoKtp:        "3171010101010001",
				Nama:         "Budi Santoso (Dummy)",
				Alamat:       "Jl. Mawar No. 12, Jakarta",
				Blok:         "A",
				Unit:         "01",
				NoTelp:       "628986937819",
				Email:        "budi.dummy@example.com",
				IDBrand:      &brandJakinet,
				TglInstalasi: &now,
			}
			if err := tx.Create(&dummyPelanggan).Error; err == nil {
				dataTeknis := domain.DataTeknis{
					PelangganID:   dummyPelanggan.ID,
					IDPelanggan:   "budi_jakinet",
					PasswordPppoe: "123456",
				}
				tx.Create(&dataTeknis)
			}
		}

		return nil
	})
}
