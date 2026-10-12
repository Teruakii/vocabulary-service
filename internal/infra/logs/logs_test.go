package logs

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGinMiddleware_DoesNotLogQueryString(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	router := gin.New()
	router.Use(GinMiddleware(slog.New(slog.NewJSONHandler(&buf, nil))))
	router.GET("/api/v1/auth/google/callback", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback?code=secret-code&state=secret-state", nil)
	router.ServeHTTP(httptest.NewRecorder(), req)

	logged := buf.String()
	assert.Contains(t, logged, `"path":"/api/v1/auth/google/callback"`)
	assert.False(t, strings.Contains(logged, "secret-code"), "auth code leaked into logs: %s", logged)
	assert.False(t, strings.Contains(logged, "secret-state"), "state leaked into logs: %s", logged)
}
