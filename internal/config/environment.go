package config

import (
	"errors"
	"os"
	"strings"
)

type Environment struct {
	PORT           string
	APP_NAME       string
	JWT_SECRET_KEY string
	APP_ENV        string

	DB_HOST     string
	DB_PORT     string
	DB_USER     string
	DB_PASSWORD string
	DB_NAME     string

	GOOGLE_OAUTH_CLIENT_ID     string
	GOOGLE_OAUTH_CLIENT_SECRET string
	GOOGLE_OAUTH_REDIRECT_URL  string

	// FRONTEND_URL is where the Google callback sends the browser after login.
	FRONTEND_URL string

	// ADMIN_EMAILS lists accounts (comma-separated) that get the admin role
	// once their email is verified.
	ADMIN_EMAILS []string

	// SMTP_HOST enables sending password reset email.
	SMTP_HOST     string
	SMTP_PORT     string
	SMTP_USERNAME string
	SMTP_PASSWORD string
	SMTP_FROM     string
}

func trimEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func (e *Environment) IsDevelopment() bool {
	if e == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(e.APP_ENV), "development")
}

func NewEnvironment() (*Environment, error) {
	env := trimEnv("APP_ENV")
	if env == "" {
		env = "production"
	}

	port := trimEnv("PORT")
	if port == "" {
		port = "8080"
	}

	dbHost := trimEnv("DB_HOST")
	if dbHost == "" {
		dbHost = "127.0.0.1"
	}

	dbPort := trimEnv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	// An empty HMAC key signs tokens anyone can forge, so refuse to start.
	jwtSecret := trimEnv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET_KEY must be set")
	}

	smtpHost := trimEnv("SMTP_HOST")
	smtpPort := trimEnv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587"
	}
	smtpFrom := trimEnv("SMTP_FROM")
	if smtpHost != "" && smtpFrom == "" {
		return nil, errors.New("SMTP_FROM must be set when SMTP_HOST is set")
	}

	return &Environment{
		PORT:           port,
		APP_NAME:       trimEnv("APP_NAME"),
		JWT_SECRET_KEY: jwtSecret,
		APP_ENV:        env,

		DB_HOST:     dbHost,
		DB_PORT:     dbPort,
		DB_USER:     trimEnv("DB_USER"),
		DB_PASSWORD: trimEnv("DB_PASSWORD"),
		DB_NAME:     trimEnv("DB_NAME"),

		GOOGLE_OAUTH_CLIENT_ID:     trimEnv("GOOGLE_OAUTH_CLIENT_ID"),
		GOOGLE_OAUTH_CLIENT_SECRET: trimEnv("GOOGLE_OAUTH_CLIENT_SECRET"),
		GOOGLE_OAUTH_REDIRECT_URL:  trimEnv("GOOGLE_OAUTH_REDIRECT_URL"),

		FRONTEND_URL: trimEnv("FRONTEND_URL"),
		ADMIN_EMAILS: splitList(trimEnv("ADMIN_EMAILS")),

		SMTP_HOST:     smtpHost,
		SMTP_PORT:     smtpPort,
		SMTP_USERNAME: trimEnv("SMTP_USERNAME"),
		SMTP_PASSWORD: trimEnv("SMTP_PASSWORD"),
		SMTP_FROM:     smtpFrom,
	}, nil
}

func splitList(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
