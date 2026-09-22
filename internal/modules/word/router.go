package word

import (
	"github.com/gin-gonic/gin"

	"vocabulary-service/internal/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, verifier middleware.TokenVerifier) {
	words := rg.Group("/words", middleware.RequireAuth(verifier))
	{
		words.POST("", h.Create)
		words.GET("", h.GetAll)
		words.GET("/search", h.GetByWord)
		words.GET("/:id", h.GetByID)
		words.PUT("/:id", h.Update)
		words.DELETE("/:id", h.Delete)
	}
}
