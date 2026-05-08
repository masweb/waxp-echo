package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type AdminConfig struct {
	DatabaseURL string
	JWTSecret   string
	ServerPort  string
	MediaDir    string
	Env         string
}

type RenderConfig struct {
	DatabaseURL string
	ServerPort  string
	MediaDir    string
	Env         string
}

func loadEnv() string {
	env := os.Getenv("ENV")
	if env == "" || env == "development" {
		_ = godotenv.Load()
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = "development"
	}
	return env
}

func loadCommon(envKeys ...struct {
	name     string
	required bool
}) (map[string]string, error) {
	vals := make(map[string]string, len(envKeys))
	for _, k := range envKeys {
		v := os.Getenv(k.name)
		if k.required && v == "" {
			return nil, fmt.Errorf("%s is required", k.name)
		}
		vals[k.name] = v
	}
	return vals, nil
}

func LoadAdmin() (*AdminConfig, error) {
	env := loadEnv()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	serverPort := os.Getenv("ADMIN_PORT")
	if serverPort == "" {
		serverPort = os.Getenv("SERVER_PORT")
	}
	if serverPort == "" {
		serverPort = ":8080"
	}

	mediaDir := os.Getenv("MEDIA_DIR")
	if mediaDir == "" {
		mediaDir = "./uploads"
	}
	mediaDir, err := filepath.Abs(mediaDir)
	if err != nil {
		return nil, fmt.Errorf("invalid MEDIA_DIR: %w", err)
	}

	return &AdminConfig{
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
		ServerPort:  serverPort,
		MediaDir:    mediaDir,
		Env:         env,
	}, nil
}

func LoadRender() (*RenderConfig, error) {
	env := loadEnv()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	serverPort := os.Getenv("RENDER_PORT")
	if serverPort == "" {
		serverPort = ":3000"
	}

	mediaDir := os.Getenv("MEDIA_DIR")
	if mediaDir == "" {
		mediaDir = "./uploads"
	}
	mediaDir, err := filepath.Abs(mediaDir)
	if err != nil {
		return nil, fmt.Errorf("invalid MEDIA_DIR: %w", err)
	}

	return &RenderConfig{
		DatabaseURL: databaseURL,
		ServerPort:  serverPort,
		MediaDir:    mediaDir,
		Env:         env,
	}, nil
}
