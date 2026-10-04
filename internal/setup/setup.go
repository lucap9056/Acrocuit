package setup

import (
	"acrocuit/internal/logs"
	"acrocuit/internal/options"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	HTTP     *HTTPConfig
	Options  *options.Options
	Logging  *logs.Config
	Warnings []string
}

type HTTPConfig struct {
	Addr         string
	AllowedHosts string
}

const (
	envAppEnv            = "APP_ENV"
	envHTTPAddr          = "HTTP_ADDR"
	envAllowedHosts      = "ALLOWED_HOSTS"
	envDatabaseDSN       = "DATABASE_DSN"
	envIdentityJWTSecret = "IDENTITY_JWT_SECRET"
	envLogLevel          = "LOG_LEVEL"
	envLogStdFormat      = "LOG_STD_FORMAT"
	envLogFilePath       = "LOG_FILE_PATH"
	envLogFileFormat     = "LOG_FILE_FORMAT"
	envLogMaxSize        = "LOG_MAX_SIZE"
	envLogMaxBackups     = "LOG_MAX_BACKUPS"
	envLogMaxAge         = "LOG_MAX_AGE"
	envLogCompress       = "LOG_COMPRESS"
)

const (
	sqliteScheme      = "sqlite://"
	defaultSQLiteDir  = "data"
	defaultSQLiteFile = "acrocuit.db"
)

func parseEnvironment(env string) options.Option {
	if options.Environment(env) == options.EnvironmentDevelopment {
		return options.WithEnvironment(options.EnvironmentDevelopment)
	}
	return options.WithEnvironment(options.EnvironmentProduction)
}

func defaultSQLitePath() string {
	executable, err := os.Executable()
	if err != nil {
		return filepath.Join(defaultSQLiteDir, defaultSQLiteFile)
	}
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}
	return filepath.Join(filepath.Dir(executable), defaultSQLiteDir, defaultSQLiteFile)
}

func parseDatabase(dsn string) (options.Option, options.Driver) {
	if dsn == "" {
		return options.WithSQLite(defaultSQLitePath()), options.DriverSQLite
	}
	if path, ok := strings.CutPrefix(dsn, sqliteScheme); ok {
		return options.WithSQLite(path), options.DriverSQLite
	}
	return options.WithPostgres(dsn), options.DriverPostgres
}

func parseAuth(driver options.Driver, identitySecret string) (options.Option, []string) {
	if driver == options.DriverSQLite {
		if identitySecret != "" {
			return options.WithSingleUser(), []string{envIdentityJWTSecret + " is ignored: SQLite always runs in single-user mode"}
		}
		return options.WithSingleUser(), nil
	}

	if identitySecret == "" {
		return options.WithForwardedHeader(), nil
	}
	return options.WithIdentitySecret(identitySecret), nil
}

func Load() *Config {
	database, driver := parseDatabase(getEnv(envDatabaseDSN, ""))
	auth, warnings := parseAuth(driver, getEnv(envIdentityJWTSecret, ""))

	return &Config{
		HTTP: &HTTPConfig{
			Addr:         getEnv(envHTTPAddr, ":8080"),
			AllowedHosts: getEnv(envAllowedHosts, "*"),
		},
		Options: options.New(
			parseEnvironment(getEnv(envAppEnv, "")),
			database,
			auth,
		),
		Logging: &logs.Config{
			Level:      getEnv(envLogLevel, "info"),
			StdFormat:  getEnv(envLogStdFormat, ""),
			FilePath:   getEnv(envLogFilePath, ""),
			FileFormat: getEnv(envLogFileFormat, ""),
			MaxSize:    getEnvInt(envLogMaxSize, 50),
			MaxBackups: getEnvInt(envLogMaxBackups, 3),
			MaxAge:     getEnvInt(envLogMaxAge, 28),
			Compress:   getEnvBool(envLogCompress, true),
		},
		Warnings: warnings,
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
