package consts

import (
	"crypto/rand"
	"encoding/hex"
	"os"
)

var (
	PORT          = env("PORT", "8080")
	ENV           = env("ENV", "debug")
	DB_URI        = env("DB_URI", "file:gmwe.db?mode=rwc")
	ACCESS_SECRET = env("ACCESS_SECRET", generateAccessSecret())
)

// env retrieves environment variable by key and replace it with defaultValue if not found
func env(key string, defaultValue string) (value string) {
	rawValue, isFound := os.LookupEnv(key)
	if !isFound {
		value = defaultValue
	} else {
		value = rawValue
	}
	return value
}

func generateAccessSecret() string {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return hex.EncodeToString(value)
}
