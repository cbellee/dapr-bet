package env

import (
	"os"
)

func GetEnvWithDefault(key, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultVal
}
