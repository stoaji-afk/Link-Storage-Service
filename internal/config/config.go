package config

import (
    "os"
    "strconv"
)

type Config struct {
    Port           string
    DBURL          string
    CacheSize      int
    ShortCodeLength int
}

func Load() *Config {
    return &Config{
        Port:           getEnv("PORT", "8080"),
        DBURL:          getEnv("DB_URL", "postgres://user:pass@localhost/links?sslmode=disable"),
        CacheSize:      getEnvAsInt("CACHE_SIZE", 1000),
        ShortCodeLength: getEnvAsInt("SHORT_CODE_LENGTH", 6),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}