package router

import (
	"net/http"

	"github.com/114514-art/hqa-PDMP/services/vault-service/internal/handler"
	"github.com/114514-art/hqa-PDMP/services/vault-service/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(db *gorm.DB, jwtSecret string, keyHandler *handler.APIKeyHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(ctx *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "error",
			})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "error",
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	})

	api := r.Group("/api/keys")
	api.Use(middleware.Auth(jwtSecret))

	api.GET("", keyHandler.List)
	api.POST("", keyHandler.Create)
	api.GET("/:id", keyHandler.Get)
	api.PATCH("/:id", keyHandler.Update)
	api.DELETE("/:id", keyHandler.Delete)
	api.POST("/:id/reveal", keyHandler.Reveal)

	return r
}
