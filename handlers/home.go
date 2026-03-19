package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"Go-learn/config"

	"github.com/gin-gonic/gin"
)

type Profile struct {
	Name        string   `json:"name"`
	Nickname    string   `json:"nickname"`
	Avatar      string   `json:"avatar"`
	Bio         string   `json:"bio"`
	Skills      []string `json:"skills"`
	Email       string   `json:"email"`
	GitHub      string   `json:"github"`
	Blog        string   `json:"blog"`
	Location    string   `json:"location"`
}

type Photo struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func LoadProfile() Profile {
	data, err := os.ReadFile(config.ProfilePath)
	if err != nil {
		return Profile{
			Name:     "个人博客",
			Nickname: "博主",
			Bio:      "欢迎来到我的个人空间",
		}
	}
	var p Profile
	json.Unmarshal(data, &p)
	return p
}

func SaveProfile(p Profile) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.ProfilePath, data, 0644)
}

func LoadPhotos() []Photo {
	data, err := os.ReadFile(config.PhotosPath)
	if err != nil {
		return []Photo{}
	}
	var photos []Photo
	json.Unmarshal(data, &photos)
	return photos
}

func SavePhotos(photos []Photo) error {
	data, err := json.MarshalIndent(photos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.PhotosPath, data, 0644)
}

// Home 首页
func Home(c *gin.Context) {
	profile := LoadProfile()
	photos := LoadPhotos()

	userIDStr, _ := c.Get("user_id")
	userID, _ := strconv.Atoi(userIDStr.(string))
	nickname, _ := c.Cookie("nickname")

	c.HTML(http.StatusOK, "home", gin.H{
		"title":    "首页",
		"profile":  profile,
		"photos":   photos,
		"userID":   uint(userID),
		"nickname": nickname,
		"isAdmin":  nickname == "admin",
	})
}

// UpdatePhotos 更新照片（仅 admin）
func UpdatePhotos(c *gin.Context) {
	nickname, _ := c.Cookie("nickname")
	if nickname != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有管理员可以管理照片"})
		return
	}

	urls := c.PostFormArray("urls[]")
	titles := c.PostFormArray("titles[]")
	descs := c.PostFormArray("descs[]")

	var photos []Photo
	for i := 0; i < len(urls); i++ {
		if urls[i] != "" {
			photos = append(photos, Photo{
				URL:         urls[i],
				Title:       titles[i],
				Description: descs[i],
			})
		}
	}

	if err := SavePhotos(photos); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}

	c.Redirect(http.StatusFound, "/")
}

// ShowProfile 显示个人信息编辑页面
func ShowProfile(c *gin.Context) {
	profile := LoadProfile()
	nickname, _ := c.Cookie("nickname")
	c.HTML(http.StatusOK, "profile", gin.H{
		"title":    "个人信息",
		"profile":  profile,
		"nickname": nickname,
		"isAdmin":  nickname == "admin",
	})
}

// UpdateProfile 更新个人信息
func UpdateProfile(c *gin.Context) {
	nickname, _ := c.Cookie("nickname")
	if nickname != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有管理员可以修改个人信息"})
		return
	}
	profile := Profile{
		Name:     c.PostForm("name"),
		Nickname: c.PostForm("nickname"),
		Avatar:   c.PostForm("avatar"),
		Bio:      c.PostForm("bio"),
		Email:    c.PostForm("email"),
		GitHub:   c.PostForm("github"),
		Blog:     c.PostForm("blog"),
		Location: c.PostForm("location"),
	}

	// 处理技能（逗号分隔）
	skillsStr := c.PostForm("skills")
	if skillsStr != "" {
		for _, s := range splitAndTrim(skillsStr, ",") {
			if s != "" {
				profile.Skills = append(profile.Skills, s)
			}
		}
	}

	// 保存并跳转
	if err := SaveProfile(profile); err != nil {
		nickname, _ := c.Cookie("nickname")
		c.HTML(http.StatusOK, "profile", gin.H{
			"title":    "编辑个人信息",
			"profile":  profile,
			"nickname": nickname,
			"error":    "保存失败，请重试",
		})
		return
	}
	c.Redirect(http.StatusFound, "/")
}

func splitAndTrim(s string, sep string) []string {
	parts := make([]string, 0)
	for _, p := range split(s, sep) {
		trimmed := trim(p)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func split(s string, sep string) []string {
	result := make([]string, 0)
	current := ""
	for _, ch := range s {
		if string(ch) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	result = append(result, current)
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
