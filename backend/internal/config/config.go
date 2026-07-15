package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL string
	RedisURL    string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	PaystackSecretKey  string
	PaystackPublicKey  string
	PaystackWebhookSec string

	CrossmintAPIKey    string
	CrossmintProjectID string
	CrossmintWebhook   string
	CrossmintEnv       string

	SMTPServer string
	SMTPPort   int
	SMTPUser   string
	SMTPPass   string
	EmailFrom  string

	CloudName   string
	CloudAPIKey string
	CloudSecret string

	Port        string
	AppEnv      string
	AppURL      string
	CORSOrigins string
}

func Load() Config {
	cfg := Config{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		PaystackSecretKey:   os.Getenv("PAYSTACK_SECRET_KEY"),
		PaystackPublicKey:   os.Getenv("PAYSTACK_PUBLIC_KEY"),
		PaystackWebhookSec:  os.Getenv("PAYSTACK_WEBHOOK_SECRET"),
		CrossmintAPIKey:     os.Getenv("CROSSMINT_API_KEY"),
		CrossmintProjectID:  os.Getenv("CROSSMINT_PROJECT_ID"),
		CrossmintWebhook:    os.Getenv("CROSSMINT_WEBHOOK_SECRET"),
		CrossmintEnv:        envOrDefault("CROSSMINT_ENV", "staging"),
		SMTPServer:          envOrDefault("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:            envInt("SMTP_PORT", 587),
		SMTPUser:            os.Getenv("SMTP_USER"),
		SMTPPass:            os.Getenv("SMTP_PASS"),
		EmailFrom:           envOrDefault("EMAIL_FROM", "Hotel Lobby <noreply@hotellobby.com>"),
		CloudName:           os.Getenv("CLOUDINARY_CLOUD_NAME"),
		CloudAPIKey:         os.Getenv("CLOUDINARY_API_KEY"),
		CloudSecret:         os.Getenv("CLOUDINARY_API_SECRET"),
		Port:                envOrDefault("PORT", "8000"),
		AppEnv:              envOrDefault("APP_ENV", "development"),
		AppURL:              envOrDefault("APP_URL", "http://localhost:8000"),
		CORSOrigins:         envOrDefault("CORS_ORIGINS", "http://localhost:5173"),
	}

	cfg.JWTAccessTTL = envDuration("JWT_ACCESS_TTL", 15*time.Minute)
	cfg.JWTRefreshTTL = envDuration("JWT_REFRESH_TTL", 7*24*time.Hour)

	return cfg
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
