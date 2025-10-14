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
	db     *sql.DB
	logger *lib.Logger
}

func NewController(db *sql.DB, logger *lib.Logger) *Controller {
	return &Controller{
		db,
		logger,
	}
}

func (c *Controller) Ping(ctx *gin.Context) {
	context, cancel := context.WithTimeout(context.TODO(), time.Second*15)
	defer cancel()
	err := c.db.PingContext(context)
	if err != nil {
		res := gin.H{
			"error": lib.ErrDatabase.Error(),
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), lib.ErrDatabase, nil, res)
		return
	}

	res := gin.H{
		"data": "pong",
	}
	ctx.JSON(200, res)
	c.logger.Info(ctx.Copy(), nil, res)
}
