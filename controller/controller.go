package controller

import (
	"database/sql"
	"net/http"

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
	err := c.db.Ping()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(200, gin.H{
		"data": "pong",
	})
}
