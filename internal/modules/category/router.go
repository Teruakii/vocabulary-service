package category

import (
	"github.com/gin-gonic/gin"

	"vocabulary-service/internal/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, verifier middleware.TokenVerifier) {
	auth := middleware.RequireAuth(verifier)

	cat := rg.Group("/categories", auth)
	cat.POST("", h.Create)
	cat.GET("", h.List)
	cat.DELETE("/:id", h.Delete)

	tags := rg.Group("/user-words/:id/tags", auth)
	tags.POST("", h.Tag)
	tags.GET("", h.ListForVocab)
	tags.DELETE("/:categoryId", h.Untag)
}
