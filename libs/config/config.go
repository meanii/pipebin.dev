package config

import (
	"log"
	"os"
	"path"
	"reflect"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadDotEnv loads a .env file if it exists.
// In container/production environments the file is absent and env vars
// are injected directly — that is not an error.
func LoadDotEnv(name, root string) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("failed to get working directory: ", err)
	}

	dotenvpath := path.Join(cwd, root, name)
	if _, err = os.Stat(dotenvpath); os.IsNotExist(err) {
		return // no .env file; rely on environment variables
	}

	if err = godotenv.Load(dotenvpath); err != nil {
		log.Fatal("failed to load .env file: ", err)
	}
}

// GetEnv get os.ENV, if not found return the fallback value
func GetEnv[T string | int](name string, defaultValue T) T {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	typeof := reflect.TypeOf(defaultValue).String()

	switch typeof {
	case "int":
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return any(intValue).(T)
	default:
		return any(value).(T)
	}
}

// MustGetEnv get os.ENV, if not found do fetal
func MustGetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("failed to get os.env: %s", name)
	}
	return value
}
