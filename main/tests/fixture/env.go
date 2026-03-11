package fixture

import "os"

// getEnv returns the value of the environment variable key, or defaultValue if not set.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
