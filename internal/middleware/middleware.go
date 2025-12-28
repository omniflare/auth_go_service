package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omniflare/auth_go_service/internal/db"
	"github.com/omniflare/auth_go_service/internal/firebase"
	"github.com/omniflare/auth_go_service/internal/models"
)

func AuthMiddleware () gin.HandlerFunc {
	return func (c *gin.Context)  {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H {
				"error" : "Authorization header required",
			})
			c.Abort()
			return 
		}
		parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
            c.Abort()
            return
        }
		idToken := parts[1]

		token, err := firebase.VerifyIDToken(c.Request.Context(), idToken)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
		var user models.User
        if err := db.DB.Where("firebase_uid = ?", token.UID).First(&user).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            c.Abort()
            return
        }

		c.Set("user", user)
		c.Next()
	}
}