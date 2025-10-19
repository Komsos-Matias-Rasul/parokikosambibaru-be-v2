package main

import (
	"fmt"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/conf"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/controller"
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

	gin.SetMode(gin.DebugMode)
	app := gin.Default()
	app.Use(corsMiddleware())
	db := lib.GetDB()
	defer db.Close()

	c := controller.NewController(db)

	app.GET("/ping", c.Ping)

	/*
		*
		*
			ZAITUN CLIENT API ROUTES
			---
	*/
	app.GET("/api/editions", c.GetAllEditions)
	app.GET("/api/editions/:editionId", c.GetEditionById)

	app.GET("/api/articles", c.GetArticlesByCategory)
	app.GET("/api/articles/:year/:editionId/:slug", c.GetArticleBySlug)
	app.GET("/api/articles/top", c.GetTopArticles)

	/*
		*
		*
			ZAITUN ADMIN API ROUTES
			---
	*/
	app.GET("/api/core/editions", c.CoreGetAllEditions)
	app.GET("/api/core/editions/:editionId/articles", c.CoreGetArticleByEdition)
	app.GET("/api/core/editions/:editionId", c.CoreGetEditionInfo)
	app.PUT("/api/core/editions/:editionId", c.CoreEditEditionInfo)
	app.POST("/api/core/edition", c.CoreCreateEdition)

	app.PUT("/api/core/articles/:articleId/archive", c.CoreArchiveArticle)
	app.DELETE("/api/core/articles/:articleId", c.CoreDeleteArticlePermanent)
	app.POST("/api/core/articles/publish/:articleId", c.CorePublishArticle)
	app.POST("/api/core/articles/create/:editionId", c.CoreCreateArticle)
	app.POST("/api/core/articles/saveDraft", c.CoreSaveDraft)
	app.POST("/api/core/articles/saveTWC", c.CoreSaveTWC)
	app.GET("/api/core/articles/:articleId", c.CoreGetArticleById)

	app.GET("/api/core/drafts", c.CoreGetDrafts)

	app.GET("/api/core/categories/by-edition/:editionId", c.GetCategoriesByEdition)
	app.GET("/api/core/categories/by-article/:articleId", c.GetCategoriesByArticle)

	app.GET("/api/core/writers", c.CoreGetAllWriters)
	app.POST("/api/core/writer", c.CoreCreateWriter)

	/*
		*
		*
			IMAGE API ROUTES
			---
	*/
	app.GET("/api/img/zaitun/editions/:year/:editionId/:fileName", c.GetZaitunCoverImg)
	app.GET("/api/img/zaitun/articles/:year/:articleId/:fileName", c.GetArticleCoverImg)
	app.GET("/api/ads/:year/:fileName", c.GetAdImage)
	app.POST("/api/core/cover/edition/save", c.CoreSaveEditionCover)
	app.POST("/api/core/cover/article/save", c.CoreSaveArticleCover)
	app.POST("/api/core/content/article/save", c.CoreSaveArticleContent)

	app.Run(fmt.Sprintf("0.0.0.0:%d", conf.SERVER_PORT))
}
