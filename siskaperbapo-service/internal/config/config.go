package config

import (
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL    string
	CloudinaryURL  string
	Port           string
	AllowedOrigins []string
	Environment    string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	ContextTimeout time.Duration
	RequestTimeout time.Duration
	RateLimitMax        int
	RateLimitExpiration time.Duration
}

func Load() *Config {
	cfg := &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		CloudinaryURL: os.Getenv("CLOUDINARY_URL"),
		Port:          os.Getenv("PORT"),
		Environment:   getEnvOrDefault("ENVIRONMENT", "development"),
		MaxConns:        20,
		MinConns:        5,
		MaxConnLifetime: 1 * time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
		ContextTimeout: 5 * time.Second,
		RequestTimeout: 10 * time.Second,
		RateLimitMax:        10,
		RateLimitExpiration: 1 * time.Minute,
	}

	originsStr := getEnvOrDefault("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001,http://localhost:5173")
	cfg.AllowedOrigins = strings.Split(originsStr, ",")
	for i := range cfg.AllowedOrigins {
		cfg.AllowedOrigins[i] = strings.TrimSpace(cfg.AllowedOrigins[i])
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL tidak ditemukan di .env file")
	}
	if cfg.Port == "" {
		cfg.Port = ":8080"
	}

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Environment) == "development"
}

func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Environment) == "production"
}
