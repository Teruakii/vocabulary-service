package auth

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"vocabulary-service/internal/infra/jwt"
	"vocabulary-service/internal/infra/mail"
	"vocabulary-service/internal/infra/oauth"
)

// InitModule wires the auth module. Accounts are admins exactly when their
// email is in adminEmails and verified; existing roles are synced to that now.
// A nil mailer turns off password reset by email.
func InitModule(db *gorm.DB, tokens *jwt.Signer, refreshTTL time.Duration, secureCookies bool, google *oauth.GoogleClient, frontendURL string, adminEmails []string, mailer mail.Mailer, rg *gin.RouterGroup) Service {
	repo := NewAuthRepoGorm(db)
	if promoted, demoted, err := repo.SyncAdmins(context.Background(), adminEmails); err != nil {
		// Starting anyway would leave admins the list no longer allows.
		log.Fatalf("could not sync admins: %v", err)
	} else {
		log.Printf("ADMIN_EMAILS sync: promoted %d, demoted %d account(s)", promoted, demoted)
	}
	service := NewService(repo, repo, repo, tokens, WithAdminEmails(adminEmails), WithPasswordReset(repo, mailer, frontendURL))
	handler := NewHandler(service, refreshTTL, secureCookies, google, frontendURL)
	RegisterRoutes(rg, handler, service)
	return service
}
