package config

import (
    "errors"
    "os"
    "strconv"
)
import "github.com/joho/godotenv"

type Config struct {
    Port         int
    PostgresDSN  string
    RedisAddr    string
    RedisPassword string
    RedisDB      int
    JWTSecret    string
}

func LoadConfig() (*Config, error) {
    _ = godotenv.Load(".env")

    port, err := strconv.Atoi(getEnv("PORT", "8080"))
    if err != nil {
        return nil, errors.New("invalid PORT value")
    }

    redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
    if err != nil {
        return nil, errors.New("invalid REDIS_DB value")
    }

    cfg := &Config{
        Port:          port,
        PostgresDSN:   getEnv("POSTGRES_DSN", "host=localhost port=5432 user=postgres password=1234 dbname=gomarket sslmode=disable"),
        RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
        RedisPassword: getEnv("REDIS_PASSWORD", ""),
        RedisDB:       redisDB,
        JWTSecret:     getEnv("JWT_SECRET", "supersecretjwtkey"),
    }
    return cfg, nil
}

func getEnv(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}
