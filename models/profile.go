package models

import (
	"time"

	"gorm.io/gorm"
)

// Profile 个人资料模型
type Profile struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100" json:"name"`
	Nickname  string         `gorm:"size:100" json:"nickname"`
	Avatar    string         `gorm:"size:255" json:"avatar"`
	Bio       string         `gorm:"type:text" json:"bio"`
	Skills    string         `gorm:"type:text" json:"skills"` // 存储为 JSON 字符串或逗号分隔
	Email     string         `gorm:"size:100" json:"email"`
	GitHub    string         `gorm:"size:255" json:"github"`
	Blog      string         `gorm:"size:255" json:"blog"`
	Location  string         `gorm:"size:100" json:"location"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Photo 精彩瞬间模型
type Photo struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	URL         string         `gorm:"size:255;not null" json:"url"`
	Title       string         `gorm:"size:100" json:"title"`
	Description string         `gorm:"size:255" json:"description"`
	Order       int            `gorm:"default:0" json:"order"` // 用于排序
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
