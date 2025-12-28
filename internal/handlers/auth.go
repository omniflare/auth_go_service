package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omniflare/auth_go_service/internal/db"
	"github.com/omniflare/auth_go_service/internal/firebase"
	"github.com/omniflare/auth_go_service/internal/models"
)

type AuthHandler struct{}

type SyncUserRequest struct {
	IDToken string `json:"idToken" binding:"required"`
}

func (h *AuthHandler) SyncUser(c *gin.Context) {
	var req SyncUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := firebase.VerifyIDToken(c.Request.Context(), req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
		return
	}
	firebaseUID := token.UID
	email := token.Claims["email"].(string)
	name := token.Claims["name"].(string)
	picture := ""
	if pic, ok := token.Claims["picture"].(string); ok {
		picture = pic
	}

	authTime := int64(0)
    if authTimeFloat, ok := token.Claims["auth_time"].(float64); ok {
        authTime = int64(authTimeFloat)
    }

	var user models.User
	result := db.DB.Where("firebase_uid = ?", firebaseUID).First(&user)
	if result.Error != nil {
		user = models.User{
			FirebaseUID: firebaseUID,
			Email:       email,
			Name:        name,
			PhotoURL:    picture,
			Role:        "user",
			Metadata: &models.UserMetadata{
				CreationTimestamp:    authTime,
				LastSignInTimestamp:  authTime,
				LastRefreshTimestamp: authTime,
			},
		}

		if err := db.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	} else {
		user.Metadata.LastSignInTimestamp = token.Claims["auth_time"].(int64)
		db.DB.Save(&user)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
	})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
