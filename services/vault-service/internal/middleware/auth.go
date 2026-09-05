package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	Email string `json:"email"`
	Type  string `json:"type"`
	jwt.RegisteredClaims
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    1005,
		"message": "未认证或token已过期",
		"data":    nil,
	})
}

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			abortUnauthorized(c)
			return
		}
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		if tokenString == "" {
			abortUnauthorized(c)
			return
		}
		claims := &AccessClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims,
			func(t *jwt.Token) (any, error) {
				if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(secret), nil
			},
		)
		if err != nil || token == nil || !token.Valid {
			abortUnauthorized(c)
			return
		}
		if claims.Subject == "" || claims.Type != "access" {
			abortUnauthorized(c)
			return
		}
		c.Set("userID", claims.Subject)
		c.Set("userEmail", claims.Email)

		c.Next()
	}
}
