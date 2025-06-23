package main

import (
	"fmt"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/conf"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/controller"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	db := lib.GetDB()
	defer db.Close()

	c := controller.NewController(db)

	router.GET("/ping", c.Ping)

	router.GET("/api/editions", c.GetAllEditions)
	router.GET("/api/editions/:id", c.GetEditionById)

	router.GET("/api/articles", c.GetArticlesByCategory)
	router.GET("/api/articles/:year/:editionId/:slug", c.GetArticleBySlug)
	router.GET("/api/articles/top", c.GetTopArticles)

	router.GET("/api/img/zaitun/editions/:year/:editionId/:fileName", c.GetZaitunCoverImg)
	router.GET("/api/img/zaitun/articles/:year/:articleId/:fileName", c.GetArticleCoverImg)
	router.GET("/api/ads/:year/:fileName", c.GetAdImage)

	// misc routes
	router.GET("/api/zaitun/current", c.GetActiveEdition)

	router.Run(fmt.Sprintf("127.0.0.1:%d", conf.SERVER_PORT))
}
