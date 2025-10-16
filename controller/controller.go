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
	db  *sql.DB
	res *lib.Responses
}

func NewController(db *sql.DB) *Controller {
	logger := lib.NewLogger()
	res := lib.NewResponses(logger)
	return &Controller{
		db,
		res,
	}
}

func (c *Controller) Ping(ctx *gin.Context) {
	context, cancel := context.WithTimeout(context.TODO(), time.Second*15)
	defer cancel()
	err := c.db.PingContext(context)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	c.res.SuccessWithStatusJSON(ctx, http.StatusOK, nil, gin.H{"message": "pong"})
}
