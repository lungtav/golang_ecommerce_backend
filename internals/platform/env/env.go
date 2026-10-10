package env

import "os"

func Get(key, fallback string) string {
	value := os.Getenv(key)

	if value != "" {
		return value
	}

	return fallback
}
