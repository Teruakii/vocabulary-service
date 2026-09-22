package category

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"vocabulary-service/internal/middleware"
)

func InitModule(db *gorm.DB, verifier middleware.TokenVerifier, rg *gin.RouterGroup) {
	repo := NewRepoGorm(db)
	service := NewService(repo)
	handler := NewHandler(service)
	RegisterRoutes(rg, handler, verifier)
}
