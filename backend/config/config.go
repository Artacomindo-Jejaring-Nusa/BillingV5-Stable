package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment                     string
	DatabaseURL                     string
	XenditCallbackTokenArtacomindo  string
	XenditCallbackTokenJelantik     string
	SecretKey                       string
	Algorithm                       string
	AccessTokenExpireMinutes        int
	XenditApiKeyJakinet             string
	XenditApiKeyJelantik            string
	XenditApiUrl                    string
	EncryptionKey                   string
	AlibabaCloudApiKey              string
	WatzapApiKey                    string
	WatzapNumberKey                 string
	WatzapAccessToken               string
	Menus                           []string
	DashboardWidgets                []string
	SystemFeatures                  []string
	DashboardWidgetPermissions      map[string][]string
}

var GlobalConfig *Config

func LoadConfig() *Config {
	// Try loading .env file from various paths
	wd, err := os.Getwd()
	if err == nil {
		envPaths := []string{
			filepath.Join(wd, ".env"),
			filepath.Join(wd, "..", ".env"),
			filepath.Join(wd, "config", ".env"),
		}
		for _, path := range envPaths {
			if _, err := os.Stat(path); err == nil {
				_ = godotenv.Load(path)
				break
			}
		}
	}

	GlobalConfig = &Config{
		Environment:                     getEnv("ENVIRONMENT", "development"),
		DatabaseURL:                     getEnv("DATABASE_URL", ""),
		XenditCallbackTokenArtacomindo:  getEnv("XENDIT_CALLBACK_TOKEN_ARTACOMINDO", "default_callback_token_artacom"),
		XenditCallbackTokenJelantik:     getEnv("XENDIT_CALLBACK_TOKEN_JELANTIK", "default_callback_token_jelantik"),
		SecretKey:                       getEnv("SECRET_KEY", "default_secret_key_change_in_production"),
		Algorithm:                       getEnv("ALGORITHM", "HS256"),
		AccessTokenExpireMinutes:        getEnvAsInt("ACCESS_TOKEN_EXPIRE_MINUTES", 120),
		XenditApiKeyJakinet:             getEnv("XENDIT_API_KEY_JAKINET", ""),
		XenditApiKeyJelantik:            getEnv("XENDIT_API_KEY_JELANTIK", ""),
		XenditApiUrl:                    getEnv("XENDIT_API_URL", "https://api.xendit.co/v2/invoices"),
		EncryptionKey:                   getEnv("ENCRYPTION_KEY", "default_encryption_key_change_in_production"),
		AlibabaCloudApiKey:              getEnv("ALIBABA_CLOUD_API_KEY", ""),
		WatzapApiKey:                    getEnv("WATZAP_API_KEY", ""),
		WatzapNumberKey:                 getEnv("WATZAP_NUMBER_KEY", ""),
		WatzapAccessToken:               getEnv("WATZAP_ACCESS_TOKEN", ""),
		
		Menus: []string{
			"Dashboard", "Pelanggan", "Langganan", "Teknis", "Paket", "Invoices", "Reports",
			"Servers", "Users", "Roles", "Permissions", "SK", "Simulasi", "Inventory",
			"Dashboard_Pelanggan", "Activity_Log", "OLT", "ODP", "Trouble_Tickets", "Diskon", "AI_Analytics",
		},
		DashboardWidgets: []string{
			"pendapatan_bulanan", "statistik_pelanggan", "statistik_server", "pelanggan_per_lokasi",
			"pelanggan_per_paket", "tren_pertumbuhan", "invoice_bulanan", "status_langganan",
			"alamat_aktif", "invoice_generation_monitor", "future_invoice_projection",
			"pelanggan_statistik_utama", "pelanggan_pendapatan_jakinet", "pelanggan_distribusi_chart",
			"pelanggan_pertumbuhan_chart", "pelanggan_status_overview_chart", "pelanggan_metrik_cepat",
			"pelanggan_tren_pendapatan_chart",
		},
		SystemFeatures: []string{
			"settings", "uploads", "traffic_monitoring",
		},
		DashboardWidgetPermissions: map[string][]string{
			"pendapatan_bulanan":          {"superadmin", "admin", "manager", "finance"},
			"invoice_bulanan":             {"superadmin", "admin", "manager", "finance"},
			"invoice_generation_monitor":  {"superadmin", "admin", "manager", "finance"},
			"future_invoice_projection":   {"superadmin", "admin", "manager", "finance"},
			"statistik_server":            {"superadmin", "admin", "noc"},
			"statistik_pelanggan":         {"superadmin", "admin", "manager", "staff", "teknisi", "noc", "finance", "bos gudang"},
			"pelanggan_per_lokasi":        {"superadmin", "admin", "manager", "staff", "teknisi", "noc", "finance", "bos gudang"},
			"pelanggan_per_paket":         {"superadmin", "admin", "manager", "staff", "teknisi", "noc", "finance", "bos gudang"},
			"tren_pertumbuhan":            {"superadmin", "admin", "manager", "finance"},
			"status_langganan":            {"superadmin", "admin", "manager", "staff", "teknisi", "noc", "finance", "bos gudang"},
			"alamat_aktif":                {"superadmin", "admin", "manager", "staff", "teknisi", "noc", "finance", "bos gudang"},
			"pelanggan_statistik_utama":   {"superadmin", "admin", "manager", "staff", "teknisi", "noc", "finance", "bos gudang", "viewer"},
			"pelanggan_pendapatan_jakinet":{"superadmin", "admin", "manager", "staff", "finance"},
			"pelanggan_distribusi_chart":  {"superadmin", "admin", "manager", "staff", "teknisi", "noc", "finance", "bos gudang", "viewer"},
			"pelanggan_pertumbuhan_chart": {"superadmin", "admin", "manager", "staff", "teknisi", "noc", "finance", "bos gudang", "viewer"},
			"pelanggan_status_overview_chart": {"superadmin", "admin", "manager", "staff", "teknisi", "noc", "finance", "bos gudang", "viewer"},
			"pelanggan_metrik_cepat":      {"superadmin", "admin", "manager", "staff", "teknisi", "noc", "finance", "bos gudang"},
			"pelanggan_tren_pendapatan_chart": {"superadmin", "admin", "manager", "staff", "finance"},
		},
	}

	if GlobalConfig.DatabaseURL == "" {
		log.Println("⚠️  DatabaseURL is not configured in environment variables or .env!")
	}

	return GlobalConfig
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func (c *Config) XENDIT_API_KEYS() map[string]string {
	return map[string]string{
		"JAKINET":  c.XenditApiKeyJakinet,
		"JELANTIK": c.XenditApiKeyJelantik,
		"ajn-01":   c.XenditApiKeyJakinet,
		"ajn-02":   c.XenditApiKeyJelantik,
		"ajn-03":   c.XenditApiKeyJakinet,
	}
}

func (c *Config) XENDIT_CALLBACK_TOKENS() map[string]string {
	return map[string]string{
		"ARTACOMINDO": c.XenditCallbackTokenArtacomindo,
		"JELANTIK":    c.XenditCallbackTokenJelantik,
	}
}

func (c *Config) CanAccessWidget(widgetName string, userRole string) bool {
	originalRole := userRole
	userRole = strings.ToLower(strings.TrimSpace(userRole))
	log.Printf("[CONFIG DEBUG] Widget: '%s', Role: '%s' (Original: '%s')", widgetName, userRole, originalRole)
	
	if userRole == "admin" || userRole == "superadmin" {
		log.Printf("[CONFIG DEBUG] Access GRANTED for '%s' to '%s' (Admin bypass)", widgetName, userRole)
		return true
	}

	if allowedRoles, exists := c.DashboardWidgetPermissions[widgetName]; exists {
		for _, role := range allowedRoles {
			if strings.ToLower(strings.TrimSpace(role)) == userRole {
				log.Printf("[CONFIG DEBUG] Access GRANTED for '%s' to '%s' (List match)", widgetName, userRole)
				return true
			}
		}
		log.Printf("[CONFIG DEBUG] Access DENIED for '%s' to '%s' (Not in: %v)", widgetName, userRole, allowedRoles)
		return false
	}
	return false
}

func (c *Config) GetUserWidgets(userRole string) []string {
	var result []string
	for _, w := range c.DashboardWidgets {
		if c.CanAccessWidget(w, userRole) {
			result = append(result, w)
		}
	}
	return result
}
