package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerAddr     string
	MySQLDSN       string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	JWTSecret      string
	JWTExpireHours int
	UploadDir      string
	UploadMaxImage int
}

func Load() *Config {
	return &Config{
		ServerAddr:     getEnv("SERVER_ADDR", ":8080"),
		MySQLDSN:       getEnv("MYSQL_DSN", "root:123456@tcp(127.0.0.1:3306)/go_chat?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:      getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvInt("REDIS_DB", 0),
		JWTSecret:      getEnv("JWT_SECRET", "change-me"),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 72),
		UploadDir:      getEnv("UPLOAD_DIR", "./data/uploads"),
		UploadMaxImage: getEnvInt("UPLOAD_MAX_IMAGE_MB", 5),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}
