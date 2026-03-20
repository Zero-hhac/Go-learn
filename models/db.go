package models

import (
	"fmt"
	"log"
	"strings"

	"Go-learn/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dbUrl := config.DATABASE_URL

	if dbUrl != "" {
		if strings.HasPrefix(dbUrl, "postgres://") || strings.HasPrefix(dbUrl, "postgresql://") {
			// 连接到 PostgreSQL (例如 Neon)
			DB, err = gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
			if err != nil {
				log.Printf("Attempted PostgreSQL connection but failed: %v", err)
				DB = nil // 确保失败后进入下一个判断
			} else {
				log.Println("Successfully connected to PostgreSQL database")
			}
		} else {
			// 尝试作为 MySQL 连接字符串 (DSN)
			DB, err = gorm.Open(mysql.Open(dbUrl), &gorm.Config{})
			if err != nil {
				log.Printf("Attempted MySQL (via URL) connection but failed: %v", err)
				DB = nil
			} else {
				log.Println("Successfully connected to MySQL database via DATABASE_URL")
			}
		}
	}

	// 如果 DATABASE_URL 连接失败或未设置，尝试使用分项配置连接 MySQL
	if DB == nil {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.DBUser,
			config.DBPassword,
			config.DBHost,
			config.DBPort,
			config.DBName,
		)
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}
		log.Println("Connected to MySQL database via individual config")
	}

	// 自动迁移表结构
	DB.AutoMigrate(&User{}, &Article{}, &Comment{}, &Profile{}, &Photo{})
}
