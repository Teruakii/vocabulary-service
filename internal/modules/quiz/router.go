package quiz

import (
	"github.com/gin-gonic/gin"

	"vocabulary-service/internal/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, tokens middleware.TokenVerifier) {
	quizzes := rg.Group("/quizzes", middleware.RequireAuth(tokens))
	{
		quizzes.POST("", h.Create)
		quizzes.GET("", h.List)
		quizzes.GET("/:id", h.GetByID)
		quizzes.POST("/:id/questions/:qid/answer", h.Answer)
		quizzes.POST("/:id/complete", h.Complete)
	}
}
