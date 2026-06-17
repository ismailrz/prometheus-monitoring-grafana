package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	Env         string
	Version     string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/appdb?sslmode=disable"),
		Env:         getEnv("ENV", "development"),
		Version:     getEnv("VERSION", "1.0.0"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
