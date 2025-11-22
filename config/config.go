package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName string
	Env         string
	HTTPAddr    string
	LogLevel    string
	LogFormat   string
	Version     string

	PasswordPepper string

	JWTAlg        string
	JWTPublicPEM  string
	JWTPrivatePEM string
	JWTSecret     string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	RedisAddr string
	RedisPass string
	RedisDB   int
}

func LoadCFG() (*Config, error) {
	cfg := &Config{
		ServiceName: getEnv("AUTH_SERVICE_NAME", "authsvc"),
		Env:         strings.ToLower(getEnv("ENV", "dev")),
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		LogLevel:    strings.ToLower(getEnv("LOG_LEVEL", "info")),
		LogFormat:   strings.ToLower(getEnv("LOG_FORMAT", "json")),
		Version:     getEnv("AUTH_VERSION", "dev"),

		PasswordPepper: getEnv("PASSWORD_PEPPER", "secret"),

		JWTAlg:        getEnv("JWT_ALG", "HS256"),
		JWTPublicPEM:  getEnv("JWT_PUBLIC_PEM", "secret"),
		JWTPrivatePEM: getEnv("JWT_PRIVATE_PEM", "secret"),
		JWTSecret:     getEnv("JWT_SECRET", "secret"),

		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBUser: getEnv("DB_USER", "postgres"),
		DBPass: getEnv("DB_PASS", "secret"),
		DBName: getEnv("DB_NAME", "postgres"),

		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass: getEnv("REDIS_PASS", "secret"),
		RedisDB:   getEnvInt("REDIS_DB", 0),
	}
	var err error
	if cfg.AccessTTL, err = parseDuration("ACCESS_TTL", "15m"); err != nil {
		return nil, err
	}
	if cfg.RefreshTTL, err = parseDuration("REFRESH_TTL", "15m"); err != nil {
		return nil, err
	}
	// Критичные проверки безопасности.
	if cfg.PasswordPepper == "" {
		return nil, errors.New("PASSWORD_PEPPER is required")
	}
	switch cfg.JWTAlg {
	case "RS256":
		if cfg.JWTPrivatePEM == "" || cfg.JWTPublicPEM == "" {
			return nil, errors.New("RS256 selected: JWT_PRIVATE_PEM and JWT_PUBLIC_PEM are required")
		}
	case "HS256":
		if cfg.JWTSecret == "" {
			return nil, errors.New("HS256 selected: JWT_SECRET is required")
		}
	default:
		return nil, fmt.Errorf("unsupported JWT_ALG=%s (use RS256 or HS256)", cfg.JWTAlg)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if value == "" {
			return fallback
		}
		result, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		return result
	}
	return fallback
}

func parseDuration(key, def string) (time.Duration, error) {
	raw := getEnv(key, def)
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %s=%q: %w", key, raw, err)
	}
	return d, nil
}
