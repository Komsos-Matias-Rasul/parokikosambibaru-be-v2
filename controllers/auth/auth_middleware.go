package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/conf"
)

// filtering jwt
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak. Token tidak ditemukan."})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode signing tidak valid")
			}

			return conf.JWT_SECRET, nil 
		})

		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid atau sudah kadaluarsa"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctx.Set("adminId", claims["id"])
		}

		ctx.Next()
	}
}