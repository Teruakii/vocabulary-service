package quiz

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"vocabulary-service/internal/middleware"
	"vocabulary-service/internal/modules/userword"
)

func InitModule(db *gorm.DB, words userword.Service, stats StatsRecorder, tokens middleware.TokenVerifier, rg *gin.RouterGroup) {
	repo := NewQuizRepoGorm(db)
	service := NewService(repo, words, stats)
	handler := NewHandler(service)
	RegisterRoutes(rg, handler, tokens)
}
