package controller

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	db *sql.DB
}

func NewController(db *sql.DB) *Controller {
	return &Controller{
		db,
	}
}

func (c *Controller) Ping(ctx *gin.Context) {
	context, cancel := context.WithTimeout(context.TODO(), time.Second*15)
	defer cancel()
	err := c.db.PingContext(context)
	if err != nil {
		lib.NewLogger(err.Error())
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	lib.NewLogger("succeed")
	ctx.JSON(200, gin.H{
		"data": "pong",
	})
}
