package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDSN            string
	DBConnectTimeout time.Duration
	DBMaxConns       int32
	DBMinConns       int32
	AppPort          string
	APIKey           string
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	HTTPIdleTimeout  time.Duration
}

func Load() (Config, error) {
	if err := loadEnvFile(); err != nil {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfg := Config{
		DBDSN:            os.Getenv("DB_DSN"),
		AppPort:          envOrDefault("APP_PORT", "8080"),
		APIKey:           os.Getenv("API_KEY"),
		DBConnectTimeout: 5 * time.Second,
		DBMaxConns:       10,
		DBMinConns:       1,
		HTTPReadTimeout:  10 * time.Second,
		HTTPWriteTimeout: 15 * time.Second,
		HTTPIdleTimeout:  60 * time.Second,
	}

	var err error
	if cfg.DBConnectTimeout, err = durationFromEnv("DB_CONNECT_TIMEOUT", cfg.DBConnectTimeout); err != nil {
		return Config{}, err
	}
	if cfg.HTTPReadTimeout, err = durationFromEnv("HTTP_READ_TIMEOUT", cfg.HTTPReadTimeout); err != nil {
		return Config{}, err
	}
	if cfg.HTTPWriteTimeout, err = durationFromEnv("HTTP_WRITE_TIMEOUT", cfg.HTTPWriteTimeout); err != nil {
		return Config{}, err
	}
	if cfg.HTTPIdleTimeout, err = durationFromEnv("HTTP_IDLE_TIMEOUT", cfg.HTTPIdleTimeout); err != nil {
		return Config{}, err
	}
	if cfg.DBMaxConns, err = int32FromEnv("DB_MAX_CONNS", cfg.DBMaxConns); err != nil {
		return Config{}, err
	}
	if cfg.DBMinConns, err = int32FromEnv("DB_MIN_CONNS", cfg.DBMinConns); err != nil {
		return Config{}, err
	}

	if cfg.DBDSN == "" {
		return Config{}, errors.New("DB_DSN is required")
	}
	if cfg.APIKey == "" {
		return Config{}, errors.New("API_KEY is required")
	}
	if cfg.DBConnectTimeout <= 0 || cfg.HTTPReadTimeout <= 0 || cfg.HTTPWriteTimeout <= 0 || cfg.HTTPIdleTimeout <= 0 {
		return Config{}, errors.New("timeouts must be greater than zero")
	}
	if cfg.DBMinConns < 0 || cfg.DBMaxConns < 1 || cfg.DBMinConns > cfg.DBMaxConns {
		return Config{}, errors.New("database pool sizes must satisfy 0 <= DB_MIN_CONNS <= DB_MAX_CONNS")
	}
	if port, parseErr := strconv.Atoi(cfg.AppPort); parseErr != nil || port < 1 || port > 65535 {
		return Config{}, errors.New("APP_PORT must be a valid TCP port")
	}

	return cfg, nil
}

func loadEnvFile() error {
	if err := godotenv.Load(); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	executablePath, err := os.Executable()
	if err != nil {
		return nil
	}
	err = godotenv.Load(filepath.Join(filepath.Dir(executablePath), ".env"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return duration, nil
}

func int32FromEnv(key string, fallback int32) (int32, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return int32(parsed), nil
}
