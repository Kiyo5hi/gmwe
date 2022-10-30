package consts

import (
	"crypto/sha256"
	"math/rand"
	"os"
)

var (
	PORT          = env("PORT", "8080")
	ENV           = env("ENV", "debug")
	DB_URI        = env("DB_URI", "root:root@tcp(mysql-dev:3306)/gmwe?charset=utf8mb4&parseTime=True&loc=Local")
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
	r := rand.Int31()
	hash := sha256.Sum256([]byte(string(r)))
	return string(hash[:])
}
