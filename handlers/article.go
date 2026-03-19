package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"Go-learn/models"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
)

// ArticleList 文章列表
func ArticleList(c *gin.Context) {
	var articles []models.Article
	models.DB.Preload("User").Order("created_at desc").Find(&articles)

	profile := LoadProfile()
	nickname, _ := c.Cookie("nickname")
	userIDStr, _ := c.Get("user_id")
	userID, _ := strconv.Atoi(userIDStr.(string))

	c.HTML(http.StatusOK, "article-list", gin.H{
		"title":    "文章列表",
		"articles": articles,
		"profile":  profile,
		"nickname": nickname,
		"isAdmin":  nickname == "admin",
		"userID":   uint(userID),
	})
}

// ArticleDetail 文章详情
func ArticleDetail(c *gin.Context) {
	id := c.Param("id")
	var article models.Article
	if err := models.DB.Preload("User").First(&article, id).Error; err != nil {
		profile := LoadProfile()
		nickname, _ := c.Cookie("nickname")
		c.HTML(http.StatusNotFound, "article-detail", gin.H{
			"title":    "文章未找到",
			"error":    "文章不存在或已被删除",
			"profile":  profile,
			"nickname": nickname,
		})
		return
	}

	// 获取评论
	var comments []models.Comment
	models.DB.Preload("User").Where("article_id = ?", article.ID).Order("created_at desc").Find(&comments)

	// 渲染 Markdown
	mdHTML := markdown.ToHTML([]byte(article.Content), nil, nil)

	profile := LoadProfile()
	nickname, _ := c.Cookie("nickname")
	userIDStr, _ := c.Get("user_id")
	userID, _ := strconv.Atoi(userIDStr.(string))

	c.HTML(http.StatusOK, "article-detail", gin.H{
		"title":    article.Title,
		"article":  article,
		"content":  template.HTML(mdHTML),
		"comments": comments,
		"profile":  profile,
		"nickname": nickname,
		"isAdmin":  nickname == "admin",
		"userID":   uint(userID),
	})
}

// ShowCreateArticle 显示写文章页面
func ShowCreateArticle(c *gin.Context) {
	profile := LoadProfile()
	nickname, _ := c.Cookie("nickname")
	c.HTML(http.StatusOK, "article-create", gin.H{
		"title":    "写文章",
		"profile":  profile,
		"nickname": nickname,
	})
}

// CreateArticle 创建文章
func CreateArticle(c *gin.Context) {
	title := strings.TrimSpace(c.PostForm("title"))
	content := strings.TrimSpace(c.PostForm("content"))

	if title == "" || content == "" {
		profile := LoadProfile()
		nickname, _ := c.Cookie("nickname")
		c.HTML(http.StatusOK, "article-create", gin.H{
			"title":    "写文章",
			"error":    "标题和内容不能为空",
			"atitle":   title,
			"content":  content,
			"profile":  profile,
			"nickname": nickname,
		})
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := strconv.Atoi(userIDStr.(string))

	article := models.Article{
		Title:   title,
		Content: content,
		UserID:  uint(userID),
	}

	if err := models.DB.Create(&article).Error; err != nil {
		profile := LoadProfile()
		nickname, _ := c.Cookie("nickname")
		c.HTML(http.StatusOK, "article-create", gin.H{
			"title":    "写文章",
			"error":    "发布失败，请重试",
			"atitle":   title,
			"content":  content,
			"profile":  profile,
			"nickname": nickname,
		})
		return
	}

	c.Redirect(http.StatusFound, "/articles/"+strconv.Itoa(int(article.ID)))
}

// DeleteArticle 删除文章
func DeleteArticle(c *gin.Context) {
	id := c.Param("id")
	userIDStr, _ := c.Get("user_id")
	userID, _ := strconv.Atoi(userIDStr.(string))
	nickname, _ := c.Cookie("nickname")

	var article models.Article
	if err := models.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章未找到"})
		return
	}

	// 权限检查：作者本人或 admin 管理员可以删除
	if article.UserID != uint(userID) && nickname != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限删除此文章"})
		return
	}

	if err := models.DB.Delete(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.Redirect(http.StatusFound, "/articles")
}

// CreateComment 提交评论
func CreateComment(c *gin.Context) {
	articleIDStr := c.PostForm("article_id")
	articleID, _ := strconv.Atoi(articleIDStr)
	content := strings.TrimSpace(c.PostForm("content"))

	if content == "" {
		c.Redirect(http.StatusFound, "/articles/"+articleIDStr)
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := strconv.Atoi(userIDStr.(string))

	comment := models.Comment{
		Content:   content,
		ArticleID: uint(articleID),
		UserID:    uint(userID),
	}

	models.DB.Create(&comment)
	c.Redirect(http.StatusFound, "/articles/"+articleIDStr)
}

// DeleteComment 删除评论
func DeleteComment(c *gin.Context) {
	id := c.Param("id")
	userIDStr, _ := c.Get("user_id")
	userID, _ := strconv.Atoi(userIDStr.(string))
	nickname, _ := c.Cookie("nickname")

	var comment models.Comment
	if err := models.DB.First(&comment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "评论未找到"})
		return
	}

	// 只有评论者、文章作者或 admin 可以删除评论
	var article models.Article
	models.DB.First(&article, comment.ArticleID)

	if comment.UserID != uint(userID) && article.UserID != uint(userID) && nickname != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限删除此评论"})
		return
	}

	models.DB.Delete(&comment)
	c.Redirect(http.StatusFound, "/articles/"+strconv.Itoa(int(comment.ArticleID)))
}

// ShowEditArticle 显示编辑文章页面
func ShowEditArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article
	if err := models.DB.First(&article, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/articles")
		return
	}

	// 权限检查
	userIDStr, _ := c.Get("user_id")
	userID, _ := strconv.Atoi(userIDStr.(string))
	nickname, _ := c.Cookie("nickname")

	if article.UserID != uint(userID) && nickname != "admin" {
		c.Redirect(http.StatusFound, "/articles/"+id)
		return
	}

	profile := LoadProfile()
	c.HTML(http.StatusOK, "article-edit", gin.H{
		"title":    "编辑文章",
		"article":  article,
		"profile":  profile,
		"nickname": nickname,
	})
}

// UpdateArticle 更新文章
func UpdateArticle(c *gin.Context) {
	id := c.Param("id")
	title := strings.TrimSpace(c.PostForm("title"))
	content := strings.TrimSpace(c.PostForm("content"))

	var article models.Article
	if err := models.DB.First(&article, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/articles")
		return
	}

	// 权限检查
	userIDStr, _ := c.Get("user_id")
	userID, _ := strconv.Atoi(userIDStr.(string))
	nickname, _ := c.Cookie("nickname")

	if article.UserID != uint(userID) && nickname != "admin" {
		c.Redirect(http.StatusFound, "/articles/"+id)
		return
	}

	if title == "" || content == "" {
		profile := LoadProfile()
		c.HTML(http.StatusOK, "article-edit", gin.H{
			"title":    "编辑文章",
			"error":    "标题和内容不能为空",
			"article":  article,
			"profile":  profile,
			"nickname": nickname,
		})
		return
	}

	models.DB.Model(&article).Updates(models.Article{
		Title:   title,
		Content: content,
	})

	c.Redirect(http.StatusFound, "/articles/"+id)
}
