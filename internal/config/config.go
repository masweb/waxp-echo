package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	ServerPort  string
	MediaDir    string
	Env         string
}

func Load() (*Config, error) {
	env := os.Getenv("ENV")
	if env == "" || env == "development" {
		_ = godotenv.Load()
		env = os.Getenv("ENV")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = ":8080"
	}

	if env == "" {
		env = "development"
	}

	mediaDir := os.Getenv("MEDIA_DIR")
	if mediaDir == "" {
		mediaDir = "./uploads"
	}
	mediaDir, err := filepath.Abs(mediaDir)
	if err != nil {
		return nil, fmt.Errorf("invalid MEDIA_DIR: %w", err)
	}

	return &Config{
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
		ServerPort:  serverPort,
		MediaDir:    mediaDir,
		Env:         env,
	}, nil
}
