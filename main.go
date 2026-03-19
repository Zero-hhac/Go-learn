package main

import (
	"html/template"

	"Go-learn/config"
	"Go-learn/handlers"
	"Go-learn/middleware"
	"Go-learn/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	models.InitDB()

	// 创建 Gin 引擎
	r := gin.Default()

	// 加载所有模板文件
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/articles/*.html"))
	r.SetHTMLTemplate(tmpl)

	// 静态文件
	r.Static("/static", "./static")

	// 公开路由（无需登录）
	r.GET("/login", handlers.ShowLogin)
	r.POST("/login", handlers.Login)
	r.GET("/register", handlers.ShowRegister)
	r.POST("/register", handlers.Register)

	// 需要登录的路由（包括首页）
	auth := r.Group("/")
	auth.Use(middleware.AuthRequired())
	{
		auth.GET("/", handlers.Home)
		auth.POST("/logout", handlers.Logout)
		auth.GET("/articles", handlers.ArticleList)
		auth.GET("/articles/create", handlers.ShowCreateArticle)
		auth.POST("/articles/create", handlers.CreateArticle)
		auth.GET("/articles/edit/:id", handlers.ShowEditArticle)
		auth.POST("/articles/edit/:id", handlers.UpdateArticle)
		auth.POST("/articles/delete/:id", handlers.DeleteArticle)
		auth.GET("/articles/:id", handlers.ArticleDetail)

		// 评论相关
		auth.POST("/comments/create", handlers.CreateComment)
		auth.POST("/comments/delete/:id", handlers.DeleteComment)

		auth.GET("/profile", handlers.ShowProfile)
		auth.POST("/profile", handlers.UpdateProfile)
		auth.POST("/photos/update", handlers.UpdatePhotos)
	}

	// 启动服务器
	println("Server is running at: http://localhost" + config.ServerPort)
	r.Run(config.ServerPort)
}
