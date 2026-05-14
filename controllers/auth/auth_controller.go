package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/conf"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)


type AuthController struct {
	db *sql.DB
	res *lib.Responses
}

func NewAuthController(db *sql.DB, res *lib.Responses) *AuthController {
	return &AuthController{
		db:  db,
		res: res,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Format email atau password tidak valid"})
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	var adminId int
	var passwordHash string

	// email
	query := `SELECT id, password_hash FROM admin_users WHERE email = ? LIMIT 1`
	err := c.db.QueryRowContext(_context, query, req.Email).Scan(&adminId, &passwordHash)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Kredensial tidak valid"})
			return
		}

		
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// pw
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Kredensial tidak valid"})
		return
	}

	// jwt 4jam expired
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    adminId,
		"email": req.Email,
		"exp":   time.Now().Add(time.Hour * 4).Unix(),
	})

	tokenString, err := token.SignedString(conf.JWT_SECRET)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"token":   tokenString,
	})
}