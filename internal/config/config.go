// Package config loads application configuration from environment variables.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	HTTP            HTTPConfig
	ShutdownTimeout time.Duration
	LogLevel        string
	AppConfig       AppConfig
}

type HTTPConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type AppConfig struct {
	CSVPath string
}

func Load(envFile ...string) Config {
	if len(envFile) == 0 {
		_ = godotenv.Load()
	} else if envFile[0] != "" {
		_ = godotenv.Overload(envFile[0])
	}

	return Config{
		Port:            getEnv("PORT", "6733"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		ShutdownTimeout: time.Duration(getEnvInt("APP_SHUTDOWN_TIMEOUT_SEC", 15)) * time.Second,
		HTTP: HTTPConfig{
			ReadTimeout:  time.Duration(getEnvInt("HTTP_READ_TIMEOUT_SEC", 10)) * time.Second,
			WriteTimeout: time.Duration(getEnvInt("HTTP_WRITE_TIMEOUT_SEC", 10)) * time.Second,
			IdleTimeout:  time.Duration(getEnvInt("HTTP_IDLE_TIMEOUT_SEC", 60)) * time.Second,
		},
		AppConfig: AppConfig{
			CSVPath: getEnv("CSV_PATH", "devices.csv"),
		},
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
