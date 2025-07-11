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
	router := gin.Default()
	router.Use(corsMiddleware())
	db := lib.GetDB()
	defer db.Close()

	c := controller.NewController(db)

	router.GET("/ping", c.Ping)

	router.GET("/api/editions", c.GetAllEditions)
	router.GET("/api/editions/:id", c.GetEditionById)

	router.GET("/api/articles", c.GetArticlesByCategory)
	router.GET("/api/articles/:year/:editionId/:slug", c.GetArticleBySlug)
	router.GET("/api/articles/top", c.GetTopArticles)

	// admin
	router.PUT("/api/core/articles/archive/:id", c.ArchiveArticle)  
	router.DELETE("/api/core/articles/delete/:id", c.DeleteArticlePermanent)  
	router.POST("/api/core/articles/publish/:id", c.PublishArticle)  
	router.POST("/api/core/articles/:editionId/create", c.CreateArticle)  
	router.POST("/api/core/articles/saveDraft", c.SaveDraft)  
	router.POST("/api/core/articles/saveTWC", c.SaveTWC) 
	router.GET("/api/core/categories", c.GetCategoriesByEdition)  


	router.GET("/api/img/zaitun/editions/:year/:editionId/:fileName", c.GetZaitunCoverImg)
	router.GET("/api/img/zaitun/articles/:year/:articleId/:fileName", c.GetArticleCoverImg)
	router.GET("/api/ads/:year/:fileName", c.GetAdImage)

	router.GET("/api/zaitun/current", c.GetActiveEdition) // deprecated

	/*Core API (admin)*/
	router.GET("/api/core/writers", c.CoreGetAllWriters)
	router.GET("/api/core/articles/:id", c.CoreGetArticleById)

	router.Run(fmt.Sprintf("127.0.0.1:%d", conf.SERVER_PORT))
}
