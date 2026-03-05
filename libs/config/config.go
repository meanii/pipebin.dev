package config

import (
	"log"
	"os"
)

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
