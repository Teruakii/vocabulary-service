package userword

import (
	"github.com/gin-gonic/gin"

	"vocabulary-service/internal/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, verifier middleware.TokenVerifier) {
	grp := rg.Group("/user-words", middleware.RequireAuth(verifier))

	grp.POST("", h.AddWord)
	grp.GET("", h.List)
	grp.GET("/due", h.Due)
	grp.GET("/:id", h.GetByID)
	grp.POST("/:id/review", h.Review)
	grp.POST("/by-word/:wordId/review", h.ReviewByWord)
	grp.DELETE("/:id", h.Remove)
}
