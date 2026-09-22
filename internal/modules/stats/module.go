package stats

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"vocabulary-service/internal/middleware"
)

func InitModule(db *gorm.DB, tokens middleware.TokenVerifier, rg *gin.RouterGroup) Service {
	repo := NewUserStatsRepoGorm(db)
	service := NewService(repo)
	handler := NewHandler(service)
	RegisterRoutes(rg, handler, tokens)
	return service
}
