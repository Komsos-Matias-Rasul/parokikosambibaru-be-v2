package main

import (
	"fmt"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/conf"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/controllers"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {

	if conf.ENV == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	if conf.ENV == "development" {
		gin.SetMode(gin.DebugMode)
	}
	app := gin.Default()
	app.Use(corsMiddleware())
	db := lib.GetDB()
	defer db.Close()

	c := controllers.NewController(db)

	app.GET("/ping", c.Ping)

	/*
		*
		*
			PROFILE API ROUTES
			---
	*/
	app.GET("/api/berita", c.Profile.GetAllBerita)
	app.GET("/api/berita/:beritaId", c.Profile.GetBeritaById)

	/*
		*
		*
			PROFILE ADMIN API ROUTES
			---
	*/
	app.POST("/api/core/berita", c.Editor.CreateBerita)

	/*
		*
		*
			ZAITUN CLIENT API ROUTES
			---
	*/
	app.GET("/api/editions", c.Zaitun.GetAllEditions)
	app.GET("/api/editions/:editionId", c.Zaitun.GetEditionById)

	app.GET("/api/articles", c.Zaitun.GetArticlesByCategory)
	app.GET("/api/articles/:year/:editionId/:slug", c.Zaitun.GetArticleBySlug)
	app.GET("/api/articles/top", c.Zaitun.GetTopArticles)

	/*
		*
		*
			ZAITUN ADMIN API ROUTES
			---
	*/
	app.POST("/api/core/edition", c.Editor.CreateEdition)
	app.GET("/api/core/editions", c.Editor.GetAllEditions)
	app.GET("/api/core/editions/:editionId/info", c.Editor.GetEditionInfo)
	app.GET("/api/core/editions/:editionId/articles", c.Editor.GetArticleByEdition)

	app.PUT("/api/core/editions/:editionId/save-info", c.Editor.EditEditionInfo)
	app.PUT("/api/core/editions/:editionId/publish", c.Editor.PublishEdition)

	app.POST("/api/core/editions/:editionId/cover", c.Image.SaveEditionCover)
	app.PUT("/api/core/editions/:editionId/cover/thumbnail", c.Image.UpdateEditionThumbnail)
	app.PUT("/api/core/editions/:editionId/cover/rename", c.Image.RenameEditionCover)

	app.POST("/api/core/article", c.Editor.CreateArticle)
	app.GET("/api/core/articles/:articleId", c.Editor.GetArticleById)
	app.GET("/api/core/articles/:articleId/info", c.Editor.GetArticleInfo)

	app.PUT("/api/core/articles/:articleId/save-info", c.Editor.SaveTWC)
	app.PUT("/api/core/articles/:articleId/save-draft", c.Editor.SaveDraft)
	app.PUT("/api/core/articles/:articleId/publish", c.Editor.PublishArticle)
	app.PUT("/api/core/articles/:articleId/archive", c.Editor.ArchiveArticle)
	app.DELETE("/api/core/articles/:articleId", c.Editor.DeleteArticlePermanent)

	app.GET("/api/core/articles/:articleId/cover", c.Image.GetArticleCoverImg)
	app.GET("/api/core/articles/:articleId/contents", c.Editor.GetArticleContent)
	app.POST("/api/core/articles/:articleId/cover", c.Image.SaveArticleCover)
	app.POST("/api/core/articles/:articleId/images", c.Image.SaveArticleImageContents)
	app.PUT("/api/core/articles/:articleId/cover/rename", c.Image.RenameArticleHeadline)
	app.PUT("/api/core/articles/:articleId/cover/thumbnail", c.Image.UpdateArticleThumbnail)

	app.GET("/api/core/drafts", c.Editor.GetDrafts)

	app.GET("/api/core/categories/by-edition/:editionId", c.Editor.GetCategoriesByEdition)
	app.GET("/api/core/categories/by-article/:articleId", c.Editor.GetCategoriesByArticle)

	app.GET("/api/core/writers", c.Editor.GetAllWriters)
	app.POST("/api/core/writer", c.Editor.CreateWriter)

	app.GET("/api/core/beritas", c.Editor.GetAllBerita)	
	app.POST("/api/core/berita/:id/cover/thumbnail", c.Editor.UpdateBeritaThumbnail)
	app.POST("api/core/berita/:id/publishing", c.Editor.UpdateBeritaPublishing)
	app.DELETE("/api/core/berita/:id", c.Editor.DeleteBeritaPermanent)
	


	/*
		*
		*
			IMAGE API ROUTES
			---
	*/
	// app.GET("/api/ads/:year/:fileName", c.GetAdImage)

	app.Run(fmt.Sprintf("0.0.0.0:%d", conf.SERVER_PORT))
}
