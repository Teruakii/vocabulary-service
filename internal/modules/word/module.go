package word

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"vocabulary-service/internal/middleware"
)

func InitModule(db *gorm.DB, verifier middleware.TokenVerifier, rg *gin.RouterGroup) Service {
	repo := NewWordRepoGorm(db)
	service := NewService(repo)
	handler := NewHandler(service)
	RegisterRoutes(rg, handler, verifier)
	return service
}
