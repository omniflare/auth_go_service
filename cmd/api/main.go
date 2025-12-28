package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/omniflare/auth_go_service/internal/config"
	"github.com/omniflare/auth_go_service/internal/db"
	"github.com/omniflare/auth_go_service/internal/firebase"
	"github.com/omniflare/auth_go_service/internal/routes"
)

func main() {
	envConfig := config.NewEnvConfig()
	db.Init(envConfig)
	firebase.Init()

	router := gin.Default()

	//CORS setup
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	routes.SetUpRoutes(router)

	port := fmt.Sprintf(":%s", envConfig.PORT)
	log.Printf("Server starting on port %s", port)
	router.Run(port)
}
