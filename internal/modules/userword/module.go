package userword

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"vocabulary-service/internal/modules/auth"
	"vocabulary-service/internal/modules/word"
)

func InitModule(db *gorm.DB, authSvc auth.Service, wordSvc word.Service, stats ActivityRecorder, rg *gin.RouterGroup) Service {
	repo := NewUserWordRepoGorm(db)
	service := NewService(repo, wordSvc, stats)
	handler := NewHandler(service)
	RegisterRoutes(rg, handler, authSvc)
	return service
}
