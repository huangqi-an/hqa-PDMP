package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(db *gorm.DB) *gin.Engine {
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
	return r
}
