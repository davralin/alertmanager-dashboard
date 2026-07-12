package config

import (
	"net"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	ListenAddr string
	Valkey     redis.Options
	StaleAfter time.Duration
}

func Load() App {
	return App{
		ListenAddr: env("LISTEN_ADDR", ":8080"),
		Valkey: redis.Options{
			Addr:     env("VALKEY_ADDR", "127.0.0.1:6379"),
			Username: os.Getenv("VALKEY_USERNAME"),
			Password: os.Getenv("VALKEY_PASSWORD"),
			DB:       envInt("VALKEY_DB", 0),
		},
		StaleAfter: envDuration("STALE_AFTER", 90*time.Minute),
	}
}

func HealthcheckURL() string {
	if value := os.Getenv("HEALTHCHECK_URL"); value != "" {
		return value
	}

	addr := env("LISTEN_ADDR", ":8080")
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "http://127.0.0.1:8080/healthz"
	}
	if host == "" || host == "::" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port) + "/healthz"
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
