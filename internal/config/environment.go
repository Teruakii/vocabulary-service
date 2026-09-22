package config

import (
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

	return &Environment{
		PORT:           port,
		APP_NAME:       trimEnv("APP_NAME"),
		JWT_SECRET_KEY: trimEnv("JWT_SECRET_KEY"),
		APP_ENV:        env,

		DB_HOST:     dbHost,
		DB_PORT:     dbPort,
		DB_USER:     trimEnv("DB_USER"),
		DB_PASSWORD: trimEnv("DB_PASSWORD"),
		DB_NAME:     trimEnv("DB_NAME"),

		GOOGLE_OAUTH_CLIENT_ID:     trimEnv("GOOGLE_OAUTH_CLIENT_ID"),
		GOOGLE_OAUTH_CLIENT_SECRET: trimEnv("GOOGLE_OAUTH_CLIENT_SECRET"),
		GOOGLE_OAUTH_REDIRECT_URL:  trimEnv("GOOGLE_OAUTH_REDIRECT_URL"),
	}, nil
}
