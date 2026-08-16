package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 应用配置，从环境变量 / .env 读取。
type Config struct {
	AppName            string
	Debug              bool
	SecretKey          string
	PanelHost          string
	PanelPort          int
	PanelBaseURL       string
	DBPath             string
	SessionDays        int
	AdminUsername      string
	AdminPassword      string
	AdminEmail         string
	SmtpHost           string
	SmtpPort           int
	SmtpUser           string
	SmtpPassword       string
	SmtpFrom           string
	SmtpUseTLS         bool
	SmtpUseSSL         bool
	DomainSuffix       string
	SigninPoints       int
	SigninStreakBonus  int
	SigninStreakDays   int
	ExchangePoints     int
	ExchangePlanDays   int
	NodeHeartbeatTO    int
	BcryptRounds       int
	RateLimitEnabled   bool
	WebDistDir         string
	StaticDir          string
	TemplatesDir       string
	LogsDir            string
}

var C *Config

// Load 加载配置。默认值对齐原 Python 实现。
func Load() *Config {
	base, _ := filepath.Abs(".")
	_ = godotenv.Load(filepath.Join(base, ".env"))
	_ = godotenv.Load()

	cfg := &Config{
		AppName:           get("APP_NAME", "WeaveNet 织网穿透"),
		Debug:             getBool("DEBUG", false),
		SecretKey:         get("SECRET_KEY", "weavenet-dev-secret-key-please-change-in-production"),
		PanelHost:         get("PANEL_HOST", "0.0.0.0"),
		PanelPort:         getInt("PANEL_PORT", 8000),
		PanelBaseURL:      get("PANEL_BASE_URL", "http://localhost:8000"),
		DBPath:            get("DATABASE_URL", ""),
		SessionDays:       getInt("SESSION_DAYS", 30),
		AdminUsername:     get("ADMIN_USERNAME", "admin"),
		AdminPassword:     get("ADMIN_PASSWORD", "admin123"),
		AdminEmail:        get("ADMIN_EMAIL", "admin@weave.test"),
		SmtpHost:          get("SMTP_HOST", ""),
		SmtpPort:          getInt("SMTP_PORT", 465),
		SmtpUser:          get("SMTP_USER", ""),
		SmtpPassword:      get("SMTP_PASSWORD", ""),
		SmtpFrom:          get("SMTP_FROM", ""),
		SmtpUseTLS:        getBool("SMTP_USE_TLS", true),
		SmtpUseSSL:        getBool("SMTP_USE_SSL", true),
		DomainSuffix:      get("DOMAIN_SUFFIX", "weave.test"),
		SigninPoints:      getInt("SIGNIN_POINTS", 10),
		SigninStreakBonus: getInt("SIGNIN_STREAK_BONUS", 30),
		SigninStreakDays:  getInt("SIGNIN_STREAK_DAYS", 7),
		ExchangePoints:    getInt("EXCHANGE_POINTS", 300),
		ExchangePlanDays:  getInt("EXCHANGE_PLAN_DAYS", 30),
		NodeHeartbeatTO:   getInt("NODE_HEARTBEAT_TIMEOUT", 60),
		BcryptRounds:      getInt("BCRYPT_ROUNDS", 12),
		RateLimitEnabled:  getBool("RATE_LIMIT_ENABLED", true),
		WebDistDir:        get("WEB_DIST_DIR", filepath.Join(base, "web", "dist")),
		StaticDir:         get("STATIC_DIR", filepath.Join(base, "static")),
		TemplatesDir:      get("TEMPLATES_DIR", filepath.Join(base, "templates")),
		LogsDir:           get("LOGS_DIR", filepath.Join(base, "logs")),
	}
	if cfg.DBPath == "" {
		cfg.DBPath = filepath.Join(base, "weavenet.db")
	} else {
		cfg.DBPath = strings.TrimPrefix(cfg.DBPath, "sqlite:///")
	}
	if cfg.BcryptRounds < 10 {
		cfg.BcryptRounds = 10
	}
	os.MkdirAll(cfg.LogsDir, 0o755)
	C = cfg
	return cfg
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getBool(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
