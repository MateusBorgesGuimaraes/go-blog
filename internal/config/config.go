package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	BaseURL     string
	ImageServerBaseUrl string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		BaseURL:     getEnv("BASE_URL", "http://localhost:5173"),
		ImageServerBaseUrl:  getEnv("IMAGE_SERVER_BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
