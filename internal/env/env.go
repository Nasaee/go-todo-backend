package env

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type AppEnv string

const (
	EnvDevelopment AppEnv = "development"
	EnvProduction  AppEnv = "production"
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
	log.Println("Environment variables loaded ✨")
}

func GetString(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func GetInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		i, err := strconv.Atoi(val)
		if err != nil {
			log.Printf("warning: env %s must be integer but got '%s', using fallback %d", key, val, fallback)
			return fallback
		}
		return i
	}
	return fallback
}
