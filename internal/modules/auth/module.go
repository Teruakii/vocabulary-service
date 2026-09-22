package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"vocabulary-service/internal/infra/jwt"
	"vocabulary-service/internal/infra/oauth"
)

func InitModule(db *gorm.DB, tokens *jwt.Signer, refreshTTL time.Duration, secureCookies bool, google *oauth.GoogleClient, rg *gin.RouterGroup) Service {
	repo := NewAuthRepoGorm(db)
	service := NewService(repo, repo, tokens)
	handler := NewHandler(service, refreshTTL, secureCookies, google)
	RegisterRoutes(rg, handler, service)
	return service
}
