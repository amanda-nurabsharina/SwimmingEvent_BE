package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv         string
	Port           string
	DBDriver       string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	APISecretKey   string
	JWTSecret      string
	AllowedOrigins string
	RateLimitMax   int
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		DBDriver:       getEnv("DB_DRIVER", "sqlite"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "swimming_event"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		APISecretKey:   getEnv("API_SECRET_KEY", "secret-swimming-api-key-2026"),
		JWTSecret:      getEnv("JWT_SECRET", "swimming-jwt-super-secret-key"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001,https://admin_masc.fourplusone.my.id,https://masc.fourplusone.my.id"),
		RateLimitMax:   60,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
