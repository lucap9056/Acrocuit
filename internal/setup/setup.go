package setup

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/logs"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment string
	HTTP        *HTTPConfig
	Database    *DatabaseConfig
	JWT         *auth.Config
	Logging     *logs.Config
}

type HTTPConfig struct {
	Addr         string
	AllowedHosts string
}

type DatabaseConfig struct {
	DSN string
}

func Load() *Config {
	return &Config{
		Environment: getEnv("APP_ENV", "production"),
		HTTP: &HTTPConfig{
			Addr:         getEnv("HTTP_ADDR", ":8080"),
			AllowedHosts: getEnv("ALLOWED_HOSTS", "*"),
		},
		Database: &DatabaseConfig{
			DSN: getEnv("DATABASE_DSN", ""),
		},
		JWT: &auth.Config{
			Issuer:             getEnv("JWT_ISSUER", "Acrocuit"),
			Audience:           getEnv("JWT_AUDIENCE", "AcrocuitClient"),
			SecretKey:          getEnv("JWT_SECRET_KEY", ""),
			AccessTokenExpiry:  time.Duration(getEnvInt("JWT_ACCESS_TOKEN_EXPIRY_MINUTES", 15)) * time.Minute,
			RefreshTokenExpiry: time.Duration(getEnvInt("JWT_REFRESH_TOKEN_EXPIRY_DAYS", 7)) * 24 * time.Hour,
		},
		Logging: &logs.Config{
			Level:      getEnv("LOG_LEVEL", "info"),
			StdFormat:  getEnv("LOG_STD_FORMAT", ""),
			FilePath:   getEnv("LOG_FILE_PATH", ""),
			FileFormat: getEnv("LOG_FILE_FORMAT", ""),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 50),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 28),
			Compress:   getEnvBool("LOG_COMPRESS", true),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
