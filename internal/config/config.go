package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBDSN         string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	AppPort       string
	StoragePath   string
}

func Load() (*Config, error) {
	cfg := &Config{
		DBDSN:         os.Getenv("DB_DSN"),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		AppPort:       os.Getenv("APP_PORT"),
		StoragePath:   os.Getenv("STORAGE_PATH"),
	}

	// Default Storage Path
	if cfg.StoragePath == "" {
		cfg.StoragePath = "/var/openbook/storage" // Default for Linux First
	}

	// Validate required fields
	if cfg.DBDSN == "" {
		return nil, fmt.Errorf("DB_DSN is required")
	}
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR is required")
	}
	if cfg.AppPort == "" {
		cfg.AppPort = "8080"
	}

	// Redis DB
	redisDBStr := os.Getenv("REDIS_DB")
	if redisDBStr == "" {
		cfg.RedisDB = 0
	} else {
		db, err := strconv.Atoi(redisDBStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
		}
		cfg.RedisDB = db
	}

	return cfg, nil
}
