package auth

import (
	"database/sql"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	db  *sql.DB
	res *lib.Responses
}

func NewAuthController(db *sql.DB, res *lib.Responses) *AuthController {
	return &AuthController{db, res}
}

func (c *AuthController) Login(ctx *gin.Context) {
	type Request struct {
		Username string `json:"username"`
		Password string `json:"password"` // don't log this
	}

	var payload Request
	if err := ctx.BindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}
	if payload.Username == "" || payload.Password == "" {
		c.res.AbortInvalidRequestBody(ctx, lib.ErrInvalidBody, "missing required field: (username, password)", nil)
		return
	}
}
