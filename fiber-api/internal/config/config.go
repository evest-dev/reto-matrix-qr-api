// Package config centraliza la carga de variables de entorno del servicio.
package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultPort               = "3000"
	defaultStatisticsAPIURL   = "http://localhost:4000"
	defaultHTTPTimeoutSeconds = 10
)

// Config son los parámetros de arranque, leídos del entorno.
type Config struct {
	// Port: env PORT, default 3000.
	Port string
	// StatisticsAPIURL es la URL base de la API de estadísticas (express-api).
	// No se hardcodea: dentro de Docker el servicio se resuelve por nombre de host.
	StatisticsAPIURL string
	// HTTPTimeout: env HTTP_TIMEOUT_SECONDS, default 10s.
	HTTPTimeout time.Duration
}

// Load lee la configuración del entorno, con valores por defecto si faltan.
func Load() Config {
	return Config{
		Port:             getEnv("PORT", defaultPort),
		StatisticsAPIURL: getEnv("STATISTICS_API_URL", defaultStatisticsAPIURL),
		HTTPTimeout:      time.Duration(getEnvInt("HTTP_TIMEOUT_SECONDS", defaultHTTPTimeoutSeconds)) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
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
