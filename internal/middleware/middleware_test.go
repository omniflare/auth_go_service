package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_NoAuthHeader(t *testing.T) {
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request, _ = http.NewRequest("GET", "/test", nil)

    AuthMiddleware()(c)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
    assert.True(t, c.IsAborted())
}

func TestAuthMiddleware_InvalidAuthHeaderFormat(t *testing.T) {
    gin.SetMode(gin.TestMode)

    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request, _ = http.NewRequest("GET", "/test", nil)
    c.Request.Header.Set("Authorization", "InvalidFormat")

    AuthMiddleware()(c)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
    assert.True(t, c.IsAborted())
}

func TestAuthMiddleware_MissingBearerPrefix(t *testing.T) {
    gin.SetMode(gin.TestMode)

    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request, _ = http.NewRequest("GET", "/test", nil)
    c.Request.Header.Set("Authorization", "SomeToken123")

    AuthMiddleware()(c)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
    assert.True(t, c.IsAborted())
}