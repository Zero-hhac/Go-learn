package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"Go-learn/config"
	"Go-learn/models"

	"github.com/gin-gonic/gin"
)

func LoadProfile() models.Profile {
	var p models.Profile
	if err := models.DB.First(&p).Error; err != nil {
		// 如果数据库中没有，尝试从 JSON 加载并同步到数据库
		data, err := os.ReadFile(config.ProfilePath)
		if err == nil {
			var temp struct {
				Name     string   `json:"name"`
				Nickname string   `json:"nickname"`
				Avatar   string   `json:"avatar"`
				Bio      string   `json:"bio"`
				Skills   []string `json:"skills"`
				Email    string   `json:"email"`
				GitHub   string   `json:"github"`
				Blog     string   `json:"blog"`
				Location string   `json:"location"`
			}
			json.Unmarshal(data, &temp)
			skillsJSON, _ := json.Marshal(temp.Skills)
			p = models.Profile{
				Name:     temp.Name,
				Nickname: temp.Nickname,
				Avatar:   temp.Avatar,
				Bio:      temp.Bio,
				Skills:   string(skillsJSON),
				Email:    temp.Email,
				GitHub:   temp.GitHub,
				Blog:     temp.Blog,
				Location: temp.Location,
			}
			models.DB.Create(&p)
		} else {
			p = models.Profile{
				Name:     "个人博客",
				Nickname: "博主",
				Bio:      "欢迎来到我的个人空间",
				Skills:   "[]",
			}
			models.DB.Create(&p)
		}
	}
	return p
}

func LoadPhotos() []models.Photo {
	var photos []models.Photo
	if err := models.DB.Order("`order` asc, created_at desc").Find(&photos).Error; err != nil || len(photos) == 0 {
		// 如果数据库中没有，尝试从 JSON 加载并同步到数据库
		data, err := os.ReadFile(config.PhotosPath)
		if err == nil {
			var temp []struct {
				URL         string `json:"url"`
				Title       string `json:"title"`
				Description string `json:"description"`
			}
			json.Unmarshal(data, &temp)
			for i, t := range temp {
				photo := models.Photo{
					URL:         t.URL,
					Title:       t.Title,
					Description: t.Description,
					Order:       i,
				}
				models.DB.Create(&photo)
				photos = append(photos, photo)
			}
		}
	}
	return photos
}

// Home 首页
func Home(c *gin.Context) {
	profile := LoadProfile()
	photos := LoadPhotos()

	// 解析 skills JSON
	var skills []string
	json.Unmarshal([]byte(profile.Skills), &skills)

	userIDStr, _ := c.Get("user_id")
	userID, _ := strconv.Atoi(userIDStr.(string))
	nickname, _ := c.Cookie("nickname")

	c.HTML(http.StatusOK, "home", gin.H{
		"title": "首页",
		"profile": gin.H{
			"Name":     profile.Name,
			"Nickname": profile.Nickname,
			"Avatar":   profile.Avatar,
			"Bio":      profile.Bio,
			"Skills":   skills,
			"Email":    profile.Email,
			"GitHub":   profile.GitHub,
			"Blog":     profile.Blog,
			"Location": profile.Location,
		},
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

	// 开启事务进行全量更新
	tx := models.DB.Begin()
	// 先删除旧数据（物理删除或软删除，这里用 Unscoped 彻底删除）
	if err := tx.Unscoped().Where("1=1").Delete(&models.Photo{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	for i := 0; i < len(urls); i++ {
		if urls[i] != "" {
			photo := models.Photo{
				URL:         urls[i],
				Title:       titles[i],
				Description: descs[i],
				Order:       i,
			}
			if err := tx.Create(&photo).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
				return
			}
		}
	}

	tx.Commit()
	c.Redirect(http.StatusFound, "/")
}

// ShowProfile 显示个人信息编辑页面
func ShowProfile(c *gin.Context) {
	profile := LoadProfile()
	// 解析 skills JSON
	var skills []string
	json.Unmarshal([]byte(profile.Skills), &skills)

	nickname, _ := c.Cookie("nickname")
	c.HTML(http.StatusOK, "profile", gin.H{
		"title": "个人信息",
		"profile": gin.H{
			"Name":     profile.Name,
			"Nickname": profile.Nickname,
			"Avatar":   profile.Avatar,
			"Bio":      profile.Bio,
			"Skills":   skills,
			"Email":    profile.Email,
			"GitHub":   profile.GitHub,
			"Blog":     profile.Blog,
			"Location": profile.Location,
		},
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

	var p models.Profile
	models.DB.First(&p)

	p.Name = c.PostForm("name")
	p.Nickname = c.PostForm("nickname")
	p.Avatar = c.PostForm("avatar")
	p.Bio = c.PostForm("bio")
	p.Email = c.PostForm("email")
	p.GitHub = c.PostForm("github")
	p.Blog = c.PostForm("blog")
	p.Location = c.PostForm("location")

	// 处理技能（逗号分隔转 JSON）
	skillsStr := c.PostForm("skills")
	var skills []string
	if skillsStr != "" {
		for _, s := range splitAndTrim(skillsStr, ",") {
			if s != "" {
				skills = append(skills, s)
			}
		}
	}
	skillsJSON, _ := json.Marshal(skills)
	p.Skills = string(skillsJSON)

	// 保存并跳转
	if err := models.DB.Save(&p).Error; err != nil {
		c.HTML(http.StatusOK, "profile", gin.H{
			"title": "编辑个人信息",
			"profile": gin.H{
				"Name":     p.Name,
				"Nickname": p.Nickname,
				"Avatar":   p.Avatar,
				"Bio":      p.Bio,
				"Skills":   skills,
				"Email":    p.Email,
				"GitHub":   p.GitHub,
				"Blog":     p.Blog,
				"Location": p.Location,
			},
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
