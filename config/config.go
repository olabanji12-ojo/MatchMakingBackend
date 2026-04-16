package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                    string
	MongoURI                string
	MongoDB                 string
	RedisAddr               string
	RedisPassword           string
	JWTSecret               string
	JWTExpiryHours          int
	BcryptCost              int
	AdminRegistrationSecret string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	return &Config{
		Port:                    getEnv("PORT", "8080"),
		MongoURI:                getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:                 getEnv("MONGO_DB", "church_match"),
		RedisAddr:               getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:           getEnv("REDIS_PASSWORD", ""),
		JWTSecret:               getEnv("JWT_SECRET", "super_secret_key"),
		JWTExpiryHours:          getEnvInt("JWT_EXPIRY_HOURS", 72),
		BcryptCost:              getEnvInt("BCRYPT_COST", 12),
		AdminRegistrationSecret: getEnv("ADMIN_REGISTRATION_SECRET", "admin_secret"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if strValue == "" {
		return fallback
	}
	value, err := strconv.Atoi(strValue)
	if err != nil {
		return fallback
	}
	return value
}
