package database

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"billing-backend/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ParseDatabaseURL converts a python-style mysql+aiomysql:// DSN into a standard Go MySQL DSN.
func ParseDatabaseURL(dbURL string) (string, error) {
	// Standard format in python env: mysql+aiomysql://user:password@host:port/database
	// Or standard: mysql://user:password@host:port/database
	if dbURL == "" {
		return "", fmt.Errorf("database URL is empty")
	}

	cleanedURL := dbURL
	if strings.Contains(cleanedURL, "://") {
		parts := strings.SplitN(cleanedURL, "://", 2)
		// replace mysql+aiomysql with mysql
		if strings.HasPrefix(parts[0], "mysql") {
			cleanedURL = "mysql://" + parts[1]
		}
	} else {
		// If it's already a DSN, return as is
		return dbURL, nil
	}

	u, err := url.Parse(cleanedURL)
	if err != nil {
		return "", err
	}

	user := u.User.Username()
	password, _ := u.User.Password()
	host := u.Host
	if host == "" {
		host = "localhost:3306"
	}
	
	// Ensure host has tcp address format
	if !strings.Contains(host, ":") {
		host = host + ":3306"
	}

	dbName := strings.TrimPrefix(u.Path, "/")
	
	// Construct MySQL DSN: username:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, dbName)
	return dsn, nil
}

// InitDatabase initializes the GORM database connection.
func InitDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn, err := ParseDatabaseURL(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	logMode := logger.Silent
	if cfg.Environment == "development" {
		logMode = logger.Info
	}

	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	}

	var db *gorm.DB
	var lastErr error

	// Retry mechanism for database connection
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(mysql.Open(dsn), gormCfg)
		if err == nil {
			break
		}
		lastErr = err
		log.Printf("⚠️ Failed to connect to database (attempt %d/5): %v. Retrying in 2s...", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if lastErr != nil && db == nil {
		return nil, fmt.Errorf("failed to connect to database after retries: %w", lastErr)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}

	// Dynamic pool configuration from config
	var poolSize, maxOverflow int
	var maxLifetime time.Duration

	if cfg.Environment == "production" {
		poolSize = 15
		maxOverflow = 25
		maxLifetime = 20 * time.Minute
	} else {
		poolSize = 20
		maxOverflow = 30
		maxLifetime = 10 * time.Minute
	}

	sqlDB.SetMaxIdleConns(poolSize)
	sqlDB.SetMaxOpenConns(poolSize + maxOverflow)
	sqlDB.SetConnMaxLifetime(maxLifetime)

	DB = db
	log.Printf("📊 Database connection pool initialized: MaxIdle=%d, MaxOpen=%d, MaxLifetime=%v", poolSize, poolSize+maxOverflow, maxLifetime)

	return db, nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// HealthCheck checks database health
func HealthCheck(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database is not initialized")
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
