package test

import "os"

const apiPath = "/api"

var (
	host          = env("ACROCUIT_HOST", "http://localhost:8080")
	identityToken = env("ACROCUIT_IDENTITY_TOKEN", "")
	accessToken   = env("ACROCUIT_ACCESS_TOKEN", "")
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
