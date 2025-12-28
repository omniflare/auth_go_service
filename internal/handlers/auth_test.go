package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omniflare/auth_go_service/internal/db"
	"github.com/omniflare/auth_go_service/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	err = db.AutoMigrate(&models.User{}, &models.UserMetadata{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	return db
}

func TestSyncUser_InvalidREquestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	invalidJSON := []byte(`{"invalid":json}`)
	c.Request, _ = http.NewRequest("POST", "/api/v1/auth/sync", bytes.NewBuffer(invalidJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &AuthHandler{}
	handler.SyncUser(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "invalid")
}

func TestSyncUser_MissingIDToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := map[string]string{}
	jsonBody, _ := json.Marshal(requestBody)
	c.Request, _ = http.NewRequest("POST", "/api/v1/auth/sync", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &AuthHandler{}
	handler.SyncUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMe_UserExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	mockUser := models.User{
		FirebaseUID: "test-uid",
		Email:       "test@example.com",
		Name:        "Test User",
		Role:        "user",
	}

	c.Set("user", mockUser)

	handler := &AuthHandler{}
	handler.GetMe(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	user := response["user"].(map[string]interface{})
	assert.Equal(t, "test@example.com", user["email"])
	assert.Equal(t, "Test User", user["name"])
}

func TestGetMe_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler := &AuthHandler{}
	handler.GetMe(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Unauthorized", response["error"])
}

func TestGetMe_WithDatabaseQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := setupTestDB(t)
	db.DB = testDB

	testUser := models.User{
		FirebaseUID: "test-uid-123",
		Email:       "dbtest@example.com",
		Name:        "DB Test User",
		Role:        "user",
	}
	testDB.Create(&testUser)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("user", testUser)

	handler := &AuthHandler{}
	handler.GetMe(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)

	user := response["user"].(map[string]interface{})
	assert.Equal(t, "dbtest@example.com", user["email"])
}
