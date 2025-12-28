package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omniflare/auth_go_service/internal/db"
	"github.com/omniflare/auth_go_service/internal/models"
	"github.com/omniflare/auth_go_service/internal/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupE2ERouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")

	db.DB = testDB

	err = testDB.AutoMigrate(&models.User{})
	require.NoError(t, err, "Failed to migrate Database")

	router := gin.Default()

	router.Use(func(ctx *gin.Context) {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}

		ctx.Next()
	})

	routes.SetUpRoutes(router)
	return router
}

func TestCompleteAuthFlow_MockUser(t *testing.T) {
	router := setupE2ERouter(t)

	t.Run("Step1_HealthCheck", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "OK", response["health"])
	})

	t.Run("Step2_ProtectedRouteWithoutAuth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Step3_SyncWithInvalidToken", func(t *testing.T) {
		requestBody := map[string]string{
			"idToken": "invalid-token-xyz",
		}
		jsonBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/sync", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.True(t, w.Code >= 400, "Should return error status")
	})
}

func TestUserJourney(t *testing.T) {
	router := setupE2ERouter(t)

	testUser := models.User{
		FirebaseUID: "e2e-test-uid-12345",
		Email:       "e2etest@example.com",
		Name:        "E2E Test User",
		PhotoURL:    "https://example.com/photo.jpg",
		Role:        "user",
		Metadata: &models.UserMetadata{
			CreationTimestamp:    time.Now().Unix(),
			LastSignInTimestamp:  time.Now().Unix(),
			LastRefreshTimestamp: time.Now().Unix(),
		},
	}

	result := db.DB.Create(&testUser)
	require.NoError(t, result.Error, "Failed to create test user")

	t.Run("Step1_VerifyUserInDatabase", func(t *testing.T) {
		var user models.User
		err := db.DB.Where("firebase_uid = ?", "e2e-test-uid-12345").First(&user).Error
		assert.NoError(t, err)
		assert.Equal(t, "e2etest@example.com", user.Email)
	})

	t.Run("Step2_AccessProtectedRoute", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer fake-token-for-e2e-test")
		router.ServeHTTP(w, req)

		assert.True(t, w.Code >= 400, "Should return error status")
	})
}

func TestConcurrentUserSync(t *testing.T) {
	router := setupE2ERouter(t)

	numConcurrent := 10
	results := make(chan int, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func(index int) {
			requestBody := map[string]string{
				"idToken": fmt.Sprintf("invalid-token-%d", index),
			}
			jsonBody, _ := json.Marshal(requestBody)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/auth/sync", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			results <- w.Code
		}(i)
	}

	for i := 0; i < numConcurrent; i++ {
		statusCode := <-results
		assert.True(t, statusCode >= 400)
	}
}

func TestDatabasePersistence(t *testing.T) {
	setupE2ERouter(t)

	users := []models.User{
		{
			FirebaseUID: "user-1",
			Email:       "user1@example.com",
			Name:        "User One",
			Role:        "user",
		},
		{
			FirebaseUID: "user-2",
			Email:       "user2@example.com",
			Name:        "User Two",
			Role:        "admin",
		},
	}

	for _, user := range users {
		db.DB.Create(&user)
	}

	var count int64
	db.DB.Model(&models.User{}).Count(&count)
	assert.Equal(t, int64(2), count)

	var user models.User
	db.DB.Where("email = ?", "user1@example.com").First(&user)
	assert.Equal(t, "User One", user.Name)
}

func TestRateLimiting(t *testing.T) {
	router := setupE2ERouter(t)

	numRequests := 100
	statusCodes := make([]int, numRequests)

	for i := 0; i < numRequests; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		router.ServeHTTP(w, req)
		statusCodes[i] = w.Code
	}

	for _, code := range statusCodes {
		assert.Equal(t, http.StatusOK, code)
	}
}

func TestErrorRecovery(t *testing.T) {
	router := setupE2ERouter(t)

	t.Run("InvalidJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/sync", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("LargePayload", func(t *testing.T) {
		largeToken := make([]byte, 1024*1024)
		for i := range largeToken {
			largeToken[i] = 'a'
		}

		requestBody := map[string]string{
			"idToken": string(largeToken),
		}
		jsonBody, _ := json.Marshal(requestBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/sync", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.True(t, w.Code >= 400)
	})

	t.Run("ServerStillResponsive", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestDataIntegrity(t *testing.T) {
	setupE2ERouter(t)

	originalUser := models.User{
		FirebaseUID: "integrity-test-uid",
		Email:       "integrity@example.com",
		Name:        "Integrity User",
		Role:        "user",
		Metadata: &models.UserMetadata{
			CreationTimestamp: time.Now().Unix(),
		},
	}

	db.DB.Create(&originalUser)

	var retrievedUser models.User
	db.DB.Where("firebase_uid = ?", "integrity-test-uid").First(&retrievedUser)

	assert.Equal(t, originalUser.Email, retrievedUser.Email)
	assert.Equal(t, originalUser.Name, retrievedUser.Name)
	assert.NotZero(t, retrievedUser.ID)
}
