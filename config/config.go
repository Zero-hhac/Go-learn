package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	ServerPort  = getEnv("PORT", ":8080")
	DBPath      = getEnv("DB_PATH", "data.db")
	ProfilePath = getEnv("PROFILE_PATH", "public/profile.json")
	PhotosPath  = getEnv("PHOTOS_PATH", "public/photos.json")

	// Database configuration
	DATABASE_URL = getEnv("DATABASE_URL", "")

	// Legacy MySQL 配置 (deprecated)
	DBHost     = getEnv("DB_HOST", "127.0.0.1")
	DBPort     = getEnv("DB_PORT", "3306")
	DBUser     = getEnv("DB_USER", "root")
	DBPassword = getEnv("DB_PASSWORD", "")
	DBName     = getEnv("DB_NAME", "blog_db")
)

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		// 云端平台通常只提供端口号如 "8080"，而 Gin 需要 ":8080"
		if key == "PORT" && value[0] != ':' {
			return ":" + value
		}
		return value
	}
	return fallback
}
