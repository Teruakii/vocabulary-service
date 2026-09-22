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

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"vocabulary-service/internal/config"
	"vocabulary-service/internal/infra/db/mysql"
	"vocabulary-service/internal/infra/jwt"
	"vocabulary-service/internal/infra/logs"
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

	models := []Model{
		&auth.User{},
		&auth.Identity{},
		&word.Word{},
		&userword.UserWord{},
		&category.Category{},
		&category.Tag{},
		&quiz.Quiz{},
		&quiz.Question{},
		&stats.UserStats{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Fatalf("failed to migrate %T: %v", model, err)
		}
	}
}

func setupRouter(env *config.Environment, db *gorm.DB, signer *jwt.Signer) *gin.Engine {
	if !env.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	logger := logs.New()

	router := gin.New()
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
	if env.GOOGLE_OAUTH_CLIENT_ID != "" && env.GOOGLE_OAUTH_CLIENT_SECRET != "" {
		googleClient = oauth.NewGoogleClient(
			env.GOOGLE_OAUTH_CLIENT_ID,
			env.GOOGLE_OAUTH_CLIENT_SECRET,
			env.GOOGLE_OAUTH_REDIRECT_URL,
		)
	}

	authService := auth.InitModule(db, signer, refreshTTL, secureCookies, googleClient, api)
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
