package main

import (
	"context"
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
	// 0. Print Banner ASCII BILLING ARTACOM V5
	logger.PrintBanner()

	// 1. Load Configurations
	cfg := config.LoadConfig()

	// 2. Initialize Fernet Encryption Service
	err := utils.InitEncryptionService(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize encryption service: %v", err)
	}
	logger.Info("Encryption service initialized successfully")

	// 3. Initialize Database Connection
	db, err := database.InitDatabase(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	
	// Initialize structured logger with DB connection
	logger.Init(db)
	logger.Info("Database connection established successfully and logger initialized")

	// Auto Migrate schemas to ensure tables are up-to-date
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
	}

	// Run Database Seeder
	if err := database.Seed(db); err != nil {
		logger.Warn("Seeding warning: %v", err)
	}

	// Clean up legacy soft-deleted customer emails & NIKs so they are released
	logger.Info("Cleaning up legacy soft-deleted customer constraints...")
	db.Exec("UPDATE pelanggans SET email = CONCAT('deleted_', UNIX_TIMESTAMP(), '_', email) WHERE deleted_at IS NOT NULL AND email NOT LIKE 'deleted_%';")
	db.Exec("UPDATE pelanggans SET no_ktp = CONCAT('deleted_', UNIX_TIMESTAMP(), '_', no_ktp) WHERE deleted_at IS NOT NULL AND no_ktp != '' AND no_ktp NOT LIKE 'deleted_%';")

	// Initialize WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()
	websocket.GlobalHub = wsHub

	router := gin.Default()

	// 5. Setup Middleware (CORS)
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve static files from uploads directory
	router.Static("/static/uploads", "./uploads")

	// Setup WebSocket Endpoint for Notifications
	router.GET("/ws/notifications", httpDelivery.HandleWebSocket)

	// 6. Setup Health Check Endpoint
	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		dbErr := database.HealthCheck(ctx)
		if dbErr != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Database connection unhealthy",
				"error":   dbErr.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"environment": cfg.Environment,
			"timestamp":   time.Now(),
		})
	})

	// 7. Setup Clean Architecture Layers (Repositories, Usecases, Handlers)
	api := router.Group("/api/v1")
	authMw := middleware.AuthMiddleware(cfg)

	// User & Auth Module
	userRepo := repository.NewUserRepository(db)
	tokenBlacklistRepo := repository.NewTokenBlacklistRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, tokenBlacklistRepo, cfg)
	httpDelivery.NewUserHandler(api, userUsecase, authMw)

	// Role & Permission Module
	roleRepo := repository.NewRoleRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	roleUsecase := usecase.NewRoleUsecase(roleRepo)
	permissionUsecase := usecase.NewPermissionUsecase(permissionRepo)
	httpDelivery.NewRoleHandler(api, roleUsecase, authMw)
	httpDelivery.NewPermissionHandler(api, permissionUsecase, authMw)

	// Layanan Module (HargaLayanan, PaketLayanan, Diskon)
	hargaLayananRepo := repository.NewHargaLayananRepository(db)
	paketLayananRepo := repository.NewPaketLayananRepository(db)
	diskonRepo := repository.NewDiskonRepository(db)
	hargaLayananUsecase := usecase.NewHargaLayananUsecase(hargaLayananRepo)
	paketLayananUsecase := usecase.NewPaketLayananUsecase(paketLayananRepo, hargaLayananRepo)
	diskonUsecase := usecase.NewDiskonUsecase(diskonRepo)
	httpDelivery.NewLayananHandler(api, hargaLayananUsecase, paketLayananUsecase, diskonUsecase, authMw)

	// Pelanggan Module
	pelangganRepo := repository.NewPelangganRepository(db)
	pelangganUsecase := usecase.NewPelangganUsecase(pelangganRepo)
	httpDelivery.NewPelangganHandler(api, pelangganUsecase, authMw)

	// Mikrotik Module
	mikrotikRepo := repository.NewMikrotikRepository(db)
	mikrotikUsecase := usecase.NewMikrotikUsecase(mikrotikRepo)
	httpDelivery.NewMikrotikHandler(api, mikrotikUsecase, authMw)

	// OLT Module
	oltRepo := repository.NewOLTRepository(db)
	oltUsecase := usecase.NewOLTUsecase(oltRepo)
	httpDelivery.NewOLTHandler(api, oltUsecase, authMw)

	// ODP Module
	odpRepo := repository.NewODPRepository(db)
	odpUsecase := usecase.NewODPUsecase(odpRepo)
	httpDelivery.NewODPHandler(api, odpUsecase, authMw)

	// DataTeknis Module
	dataTeknisRepo := repository.NewDataTeknisRepository(db)
	dataTeknisUsecase := usecase.NewDataTeknisUsecase(dataTeknisRepo, mikrotikRepo, pelangganRepo, paketLayananRepo)
	httpDelivery.NewDataTeknisHandler(api, dataTeknisUsecase, authMw)

	systemRepo := repository.NewSystemRepository(db)
	systemUsecase := usecase.NewSystemUsecase(systemRepo)

	// Billing Module
	invoiceRepo := repository.NewInvoiceRepository(db)
	langgananRepo := repository.NewLanggananRepository(db)
	billingUsecase := usecase.NewBillingUsecase(invoiceRepo, langgananRepo, pelangganRepo, paketLayananRepo, hargaLayananRepo, dataTeknisRepo, mikrotikRepo, diskonRepo, systemRepo, cfg)

	// Initialize Scheduler Manager
	schedulerMgr := scheduler.NewSchedulerManager(db, systemUsecase, billingUsecase)

	httpDelivery.NewBillingHandler(api, billingUsecase, authMw)
	httpDelivery.NewSystemHandler(api, systemUsecase, schedulerMgr, authMw)

	// Inventory Module
	inventoryRepo := repository.NewInventoryRepository(db)
	inventoryUsecase := usecase.NewInventoryUsecase(inventoryRepo, pelangganRepo, systemRepo)
	httpDelivery.NewInventoryHandler(api, inventoryUsecase, authMw)

	// Trouble Ticket Module
	troubleTicketRepo := repository.NewTroubleTicketRepository(db)
	troubleTicketUsecase := usecase.NewTroubleTicketUsecase(troubleTicketRepo, systemRepo, cfg)
	httpDelivery.NewTroubleTicketHandler(api, troubleTicketUsecase, authMw)

	// Dashboard Module
	dashboardRepo := repository.NewDashboardRepository(db)
	dashboardUsecase := usecase.NewDashboardUsecase(dashboardRepo, cfg)
	httpDelivery.NewDashboardHandler(api, dashboardUsecase, userUsecase, authMw)
	httpDelivery.NewDashboardPelangganHandler(api, dashboardRepo, authMw)

	// Notification Module
	httpDelivery.NewNotificationHandler(api, authMw)


	// 8. Start Cron Scheduler
	schedulerMgr.Start(context.Background())

	// 9. Start Server with Graceful Shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

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

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Stop cron scheduler
	logger.Info("Stopping cron scheduler...")
	schedulerMgr.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	logger.Info("Server exiting")
}
