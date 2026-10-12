package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	// Embedded zoneinfo, so TZ works in an image without OS zone files.
	_ "time/tzdata"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"vocabulary-service/internal/config"
	"vocabulary-service/internal/infra/db/mysql"
	"vocabulary-service/internal/infra/jwt"
	"vocabulary-service/internal/infra/logs"
	"vocabulary-service/internal/infra/mail"
	"vocabulary-service/internal/infra/oauth"

	"vocabulary-service/internal/modules/auth"
	"vocabulary-service/internal/modules/category"
	"vocabulary-service/internal/modules/quiz"
	"vocabulary-service/internal/modules/stats"
	"vocabulary-service/internal/modules/userword"
	"vocabulary-service/internal/modules/word"
)

const Version = "0.0.1"

func main() {
	_ = godotenv.Load()

	env, err := config.NewEnvironment()
	if err != nil {
		log.Fatalf("failed to load environment: %v", err)
	}

	db, err := mysql.NewClient(mysql.Config{
		Host:            env.DB_HOST,
		Port:            env.DB_PORT,
		User:            env.DB_USER,
		Password:        env.DB_PASSWORD,
		DBName:          env.DB_NAME,
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
	})
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	runMigrations(db)

	signer := jwt.NewSigner(jwt.Config{

		AccessSecret:  []byte(env.JWT_SECRET_KEY),
		RefreshSecret: []byte(env.JWT_SECRET_KEY),
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
		Issuer:        env.APP_NAME,
	})

	router := setupRouter(env, db, signer)

	srv := &http.Server{
		Addr:         ":" + env.PORT,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s (v%s)", env.PORT, Version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	waitForShutdown(srv, db)
}

func runMigrations(db *gorm.DB) {
	type Model interface {
		TableName() string
	}

	if err := dedupeUserWords(db); err != nil {
		log.Fatalf("failed to merge duplicate user words: %v", err)
	}

	models := []Model{
		&auth.User{},
		&auth.Identity{},
		&auth.Session{},
		&auth.PasswordReset{},
		&word.Word{},
		&userword.UserWord{},
		&category.Category{},
		&category.Tag{},
		&quiz.Quiz{},
		&quiz.Question{},
		&stats.UserStats{},
		&stats.DailyActivity{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Fatalf("failed to migrate %T: %v", model, err)
		}
	}
}

// dedupeUserWords merges duplicate (user_id, word_id) rows that older builds
// allowed, so AutoMigrate can add the unique index. The oldest row of each
// group is kept and the duplicates' category tags move onto it.
func dedupeUserWords(db *gorm.DB) error {
	if !db.Migrator().HasTable(&userword.UserWord{}) {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var dups []struct {
			DupID  uint64
			KeepID uint64
		}
		if err := tx.Raw(`SELECT d.user_word_id AS dup_id, k.keep_id
			FROM user_word d
			JOIN (SELECT user_id, word_id, MIN(user_word_id) AS keep_id
			      FROM user_word GROUP BY user_id, word_id HAVING COUNT(*) > 1) k
			  ON k.user_id = d.user_id AND k.word_id = d.word_id
			WHERE d.user_word_id <> k.keep_id`).Scan(&dups).Error; err != nil {
			return err
		}
		hasTags := tx.Migrator().HasTable(&category.Tag{})
		for _, d := range dups {
			if hasTags {
				// IGNORE skips tags the kept row already has; the delete below
				// drops those leftovers.
				if err := tx.Exec("UPDATE IGNORE vocab_tags SET vocab_id = ? WHERE vocab_id = ?", d.KeepID, d.DupID).Error; err != nil {
					return err
				}
				if err := tx.Exec("DELETE FROM vocab_tags WHERE vocab_id = ?", d.DupID).Error; err != nil {
					return err
				}
			}
			if err := tx.Exec("DELETE FROM user_word WHERE user_word_id = ?", d.DupID).Error; err != nil {
				return err
			}
		}
		if len(dups) > 0 {
			log.Printf("merged %d duplicate user word(s)", len(dups))
		}
		return nil
	})
}

func setupRouter(env *config.Environment, db *gorm.DB, signer *jwt.Signer) *gin.Engine {
	if !env.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	logger := logs.New()

	router := gin.New()
	// Use the real TCP peer as the client IP. Trusting X-Forwarded-For from
	// anyone would let a client fake its IP and dodge the login limiter.
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatalf("set trusted proxies: %v", err)
	}
	router.Use(gin.Recovery(), logs.GinMiddleware(logger))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": Version,
		})
	})

	api := router.Group("/api/v1")

	refreshTTL := 7 * 24 * time.Hour
	secureCookies := !env.IsDevelopment()

	var googleClient *oauth.GoogleClient
	if env.GOOGLE_OAUTH_CLIENT_ID != "" && env.GOOGLE_OAUTH_CLIENT_SECRET != "" && env.GOOGLE_OAUTH_REDIRECT_URL != "" {
		googleClient = oauth.NewGoogleClient(
			env.GOOGLE_OAUTH_CLIENT_ID,
			env.GOOGLE_OAUTH_CLIENT_SECRET,
			env.GOOGLE_OAUTH_REDIRECT_URL,
		)
		log.Printf("google oauth enabled (redirect: %s)", env.GOOGLE_OAUTH_REDIRECT_URL)
	} else {
		log.Println("google oauth disabled: set GOOGLE_OAUTH_CLIENT_ID, GOOGLE_OAUTH_CLIENT_SECRET and GOOGLE_OAUTH_REDIRECT_URL")
	}

	var mailer mail.Mailer
	switch {
	case env.SMTP_HOST != "":
		mailer = mail.NewSMTP(env.SMTP_HOST, env.SMTP_PORT, env.SMTP_USERNAME, env.SMTP_PASSWORD, env.SMTP_FROM)
		log.Printf("password reset email via %s:%s", env.SMTP_HOST, env.SMTP_PORT)
	case env.IsDevelopment():
		mailer = mail.Log{}
		log.Println("password reset email is logged, not sent (development)")
	default:
		log.Println("password reset disabled: set SMTP_HOST and SMTP_FROM")
	}

	authService := auth.InitModule(db, signer, refreshTTL, secureCookies, googleClient, env.FRONTEND_URL, env.ADMIN_EMAILS, mailer, api)
	wordService := word.InitModule(db, authService, api)
	category.InitModule(db, signer, api)
	statsService := stats.InitModule(db, signer, api)
	userwordService := userword.InitModule(db, authService, wordService, statsService, api)
	quiz.InitModule(db, userwordService, statsService, signer, api)

	return router
}

func waitForShutdown(srv *http.Server, db *gorm.DB) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}

	log.Println("server exited cleanly")
}
