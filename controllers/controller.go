package controllers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	a "github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/controllers/auth"
	e "github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/controllers/editor"
	i "github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/controllers/image"
	p "github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/controllers/profile"
	z "github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/controllers/zaitun"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	db      *sql.DB
	res     *lib.Responses
	Profile *p.ProfileController
	Zaitun  *z.ZaitunController
	Editor  *e.EditorController
	Image   *i.ImageController
	Auth    *a.AuthController
}

func NewController(db *sql.DB) *Controller {
	logger := lib.NewLogger()
	res := lib.NewResponses(logger)
	profile := p.NewProfileController(db, res)
	zaitun := z.NewZaitunController(db, res)
	editor := e.NewEditorController(db, res)
	image := i.NewImageController(db, res)
	auth := a.NewAuthController(db, res)
	return &Controller{
		db,
		res,
		profile,
		zaitun,
		editor,
		image,
		auth,
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
