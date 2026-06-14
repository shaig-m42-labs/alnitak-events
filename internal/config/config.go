package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	NATSURL             string
	MaxDeliveryAttempts int
}

func Load() Config {
	return Config{
		Port:                env("PORT", "8083"),
		DatabaseURL:         env("DATABASE_URL", "postgres://orion:orion@localhost:5435/orion_events?sslmode=disable"),
		NATSURL:             env("NATS_URL", "nats://localhost:4222"),
		MaxDeliveryAttempts: envInt("MAX_DELIVERY_ATTEMPTS", 3),
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
