package config

import (
	"os"
)

var (
	ServerPort  = getEnv("PORT", ":8080")
	DBPath      = getEnv("DB_PATH", "data.db")
	ProfilePath = getEnv("PROFILE_PATH", "public/profile.json")
	PhotosPath  = getEnv("PHOTOS_PATH", "public/photos.json")
)

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
