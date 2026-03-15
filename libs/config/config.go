package config

import (
	"log"
	"os"
	"path"
	"reflect"
	"strconv"

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
