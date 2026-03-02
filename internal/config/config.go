package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Ubah dari DB_URL jadi DBURL
	DBURL string
	// Ubah dari JWT_SECRET jadi JWTSecret
	JWTSecret string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		DBURL:     os.Getenv("DB_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
	}
}
