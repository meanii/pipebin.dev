package config

import (
	"log"
	"os"
	"path"

	"github.com/joho/godotenv"
)

func LoadDotEnv(name, root string) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Error failed to get CWD")
	}

	dotenvpath := path.Join(cwd, root, name)
	err = godotenv.Load(dotenvpath)
	if err != nil {
		log.Fatal("Error loading .env file, ", dotenvpath)
	}
}

// GetEnv get os.ENV, if not found return the fallback value
func GetEnv(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}

// MustGetEnv get os.ENV, if not found do fetal
func MustGetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("failed to get os.env: %s", name)
	}
	return value
}
