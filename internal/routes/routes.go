package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omniflare/auth_go_service/internal/handlers"
	"github.com/omniflare/auth_go_service/internal/middleware"
)

func SetUpRoutes(router *gin.Engine) {
	authHandler := &handlers.AuthHandler{}
	api := router.Group("/api/v1")
	{
		api.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"health": "OK", "status": "ON"})
		})

		api.POST("/auth/sync", authHandler.SyncUser)
	}

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/auth/me", authHandler.GetMe)
	}
}
