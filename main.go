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

	router.Run(fmt.Sprintf("127.0.0.1:%d", conf.SERVER_PORT))
}
