package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"billing-backend/config"
	httpDelivery "billing-backend/internal/delivery/http"
	"billing-backend/internal/domain"
	"billing-backend/internal/middleware"
	"billing-backend/internal/repository"
	"billing-backend/internal/scheduler"
	"billing-backend/internal/usecase"
	"billing-backend/internal/websocket"
	"billing-backend/pkg/database"
	"billing-backend/pkg/logger"
	"billing-backend/pkg/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 0. Print Banner
	logger.PrintBanner()

	// 1. Load Configurations
	cfg := config.LoadConfig()

	// 2. Initialize Fernet Encryption Service
	err := utils.InitEncryptionService(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize encryption service: %v", err)
	}

	// 3. Initialize Database Connection
	db, err := database.InitDatabase(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// 3.1. Ensure core system tables and columns exist immediately
	// This fixes errors where GORM expects certain columns but legacy dump doesn't have them
	db.Exec("CREATE TABLE IF NOT EXISTS system_logs (id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, timestamp DATETIME(6), level VARCHAR(50), message TEXT);")
	
	ensureDeletedAt := func(tableName string) {
		var columnCount int
		// Check if column exists in the current database
		err := db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = 'deleted_at'", tableName).Scan(&columnCount).Error
		if err == nil && columnCount == 0 {
			log.Printf("[DB FIX] Adding missing 'deleted_at' column to '%s' table...", tableName)
			db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN deleted_at DATETIME NULL DEFAULT NULL;", tableName))
			db.Exec(fmt.Sprintf("CREATE INDEX idx_%s_deleted_at ON %s (deleted_at);", tableName, tableName))
		}
	}

	// Tables that use soft delete (gorm.DeletedAt) in the new backend
	ensureDeletedAt("pelanggan")
	ensureDeletedAt("invoices")
	ensureDeletedAt("mikrotik_servers")
	ensureDeletedAt("trouble_ticket")
	ensureDeletedAt("action_taken")
	ensureDeletedAt("inventory_items")
	ensureDeletedAt("users")
	ensureDeletedAt("langganan")
	ensureDeletedAt("odp")
	ensureDeletedAt("olt")
	ensureDeletedAt("diskon")
	ensureDeletedAt("paket_layanan")
	ensureDeletedAt("harga_layanan")
	ensureDeletedAt("syarat_ketentuan")
	ensureDeletedAt("activity_logs")
	ensureDeletedAt("payment_callback_logs")

	// Initialize structured logger
	logger.Init(db)
	logger.Info("Database connection established successfully")

	// 4. Auto Migrate schemas
	// We disable foreign key checks during migration to avoid issues with legacy type mismatches
	db.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	err = db.AutoMigrate(
		&domain.Role{},
		&domain.User{},
		&domain.HargaLayanan{},
		&domain.PaketLayanan{},
		&domain.MikrotikServer{},
		&domain.OLT{},
		&domain.ODP{},
		&domain.Pelanggan{},
		&domain.DataTeknis{},
		&domain.Diskon{},
		&domain.Langganan{},
		&domain.Invoice{},
		&domain.InvoiceArchive{},
		&domain.PaymentCallbackLog{},
		&domain.TrafficHistory{},
		&domain.ActivityLog{},
		&domain.SystemLog{},
		&domain.SystemSetting{},
		&domain.SyaratKetentuan{},
		&domain.TokenBlacklist{},
		&domain.InventoryItemType{},
		&domain.InventoryStatus{},
		&domain.InventoryItem{},
		&domain.InventoryHistory{},
		&domain.TroubleTicket{},
		&domain.TicketHistory{},
		&domain.ActionTaken{},
	)
	db.Exec("SET FOREIGN_KEY_CHECKS = 1;")
	if err != nil {
		logger.Warn("AutoMigration warning: %v", err)
	} else {
		logger.Info("Database AutoMigration completed successfully")
	}

	// Standardize legacy data
	logger.Info("Standardizing legacy data...")
	db.Exec("UPDATE invoices SET status_invoice = 'Expired' WHERE status_invoice = 'Kadaluarsa';")
	db.Exec("UPDATE invoices_archive SET status_invoice = 'Expired' WHERE status_invoice = 'Kadaluarsa';")
	db.Exec("UPDATE pelanggan SET email = CONCAT('deleted_', UNIX_TIMESTAMP(), '_', email) WHERE deleted_at IS NOT NULL AND email NOT LIKE 'deleted_%';")
	db.Exec("UPDATE pelanggan SET no_ktp = CONCAT('deleted_', UNIX_TIMESTAMP(), '_', no_ktp) WHERE deleted_at IS NOT NULL AND no_ktp != '' AND no_ktp NOT LIKE 'deleted_%';")

	// Run Database Seeder
	if err := database.Seed(db); err != nil {
		logger.Warn("Seeding warning: %v", err)
	}

	// Initialize WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()
	websocket.GlobalHub = wsHub

	// Initialize Redis pub/sub for cross-instance WebSocket notifications
	websocket.InitRedis()

	router := gin.Default()

	// 5. Setup Middleware (CORS)
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Static("/static/uploads", "./uploads")
	router.GET("/ws/notifications", httpDelivery.HandleWebSocket)

	// 6. Setup Health Check
	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := database.HealthCheck(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "environment": cfg.Environment, "timestamp": time.Now()})
	})

	// 7. Setup Clean Architecture Layers
	api := router.Group("/api/v1")
	authMw := middleware.AuthMiddleware(cfg)

	// User & Auth
	userRepo := repository.NewUserRepository(db)
	tokenBlacklistRepo := repository.NewTokenBlacklistRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, tokenBlacklistRepo, cfg)
	httpDelivery.NewUserHandler(api, userUsecase, authMw)

	// Role & Permission
	roleRepo := repository.NewRoleRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	roleUsecase := usecase.NewRoleUsecase(roleRepo)
	permissionUsecase := usecase.NewPermissionUsecase(permissionRepo)
	httpDelivery.NewRoleHandler(api, roleUsecase, authMw)
	httpDelivery.NewPermissionHandler(api, permissionUsecase, authMw)

	// Layanan
	hargaLayananRepo := repository.NewHargaLayananRepository(db)
	paketLayananRepo := repository.NewPaketLayananRepository(db)
	diskonRepo := repository.NewDiskonRepository(db)
	hargaLayananUsecase := usecase.NewHargaLayananUsecase(hargaLayananRepo)
	paketLayananUsecase := usecase.NewPaketLayananUsecase(paketLayananRepo, hargaLayananRepo)
	diskonUsecase := usecase.NewDiskonUsecase(diskonRepo)
	httpDelivery.NewLayananHandler(api, hargaLayananUsecase, paketLayananUsecase, diskonUsecase, authMw)

	// Pelanggan
	pelangganRepo := repository.NewPelangganRepository(db)
	pelangganUsecase := usecase.NewPelangganUsecase(pelangganRepo)
	httpDelivery.NewPelangganHandler(api, pelangganUsecase, authMw)

	// Mikrotik
	mikrotikRepo := repository.NewMikrotikRepository(db)
	mikrotikUsecase := usecase.NewMikrotikUsecase(mikrotikRepo)
	httpDelivery.NewMikrotikHandler(api, mikrotikUsecase, authMw)

	// OLT
	oltRepo := repository.NewOLTRepository(db)
	oltUsecase := usecase.NewOLTUsecase(oltRepo)
	httpDelivery.NewOLTHandler(api, oltUsecase, authMw)

	// ODP
	odpRepo := repository.NewODPRepository(db)
	odpUsecase := usecase.NewODPUsecase(odpRepo)
	httpDelivery.NewODPHandler(api, odpUsecase, authMw)

	// DataTeknis
	dataTeknisRepo := repository.NewDataTeknisRepository(db)
	dataTeknisUsecase := usecase.NewDataTeknisUsecase(dataTeknisRepo, mikrotikRepo, pelangganRepo, paketLayananRepo)
	httpDelivery.NewDataTeknisHandler(api, dataTeknisUsecase, authMw)

	systemRepo := repository.NewSystemRepository(db)
	systemUsecase := usecase.NewSystemUsecase(systemRepo)

	// Billing
	invoiceRepo := repository.NewInvoiceRepository(db)
	langgananRepo := repository.NewLanggananRepository(db)
	billingUsecase := usecase.NewBillingUsecase(invoiceRepo, langgananRepo, pelangganRepo, paketLayananRepo, hargaLayananRepo, dataTeknisRepo, mikrotikRepo, diskonRepo, systemRepo, cfg)

	// Scheduler
	schedulerMgr := scheduler.NewSchedulerManager(db, systemUsecase, billingUsecase)
	httpDelivery.NewBillingHandler(api, billingUsecase, authMw)
	httpDelivery.NewSystemHandler(api, systemUsecase, schedulerMgr, authMw)

	// Inventory
	inventoryRepo := repository.NewInventoryRepository(db)
	inventoryUsecase := usecase.NewInventoryUsecase(inventoryRepo, pelangganRepo, systemRepo)
	httpDelivery.NewInventoryHandler(api, inventoryUsecase, authMw)

	// Trouble Ticket
	troubleTicketRepo := repository.NewTroubleTicketRepository(db)
	troubleTicketUsecase := usecase.NewTroubleTicketUsecase(troubleTicketRepo, systemRepo, cfg)
	httpDelivery.NewTroubleTicketHandler(api, troubleTicketUsecase, authMw)

	// Dashboard
	dashboardRepo := repository.NewDashboardRepository(db)
	dashboardUsecase := usecase.NewDashboardUsecase(dashboardRepo, cfg)
	httpDelivery.NewDashboardHandler(api, dashboardUsecase, userUsecase, authMw)
	httpDelivery.NewDashboardPelangganHandler(api, dashboardRepo, authMw)

	// Notifications
	httpDelivery.NewNotificationHandler(api, authMw)

	// 8. Start Cron Scheduler
	schedulerMgr.Start(context.Background())

	// 9. Start Server
	port := os.Getenv("PORT")
	if port == "" { port = "8000" }

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		logger.Info("Server is running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")
	schedulerMgr.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
		os.Exit(1)
	}
	logger.Info("Server exiting")
}
