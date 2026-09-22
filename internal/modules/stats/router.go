package stats

import (
	"github.com/gin-gonic/gin"

	"vocabulary-service/internal/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, tokens middleware.TokenVerifier) {
	s := rg.Group("/stats", middleware.RequireAuth(tokens))
	{
		s.GET("/me", h.Me)
	}
}
