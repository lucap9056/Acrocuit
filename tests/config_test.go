package test

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

const (
	apiPath           = "/api"
	identityTokenType = "identity+jwt"
)

var (
	host           = env("ACROCUIT_HOST", "http://localhost:8080")
	userEmail      = env("ACROCUIT_USER_EMAIL", "tester@localhost")
	identitySecret = env("ACROCUIT_IDENTITY_SECRET", "")
	accessToken    = env("ACROCUIT_ACCESS_TOKEN", "")
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func identityToken(t *testing.T) string {
	t.Helper()
	if token := os.Getenv("ACROCUIT_IDENTITY_TOKEN"); token != "" {
		return token
	}
	if identitySecret == "" {
		return ""
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userEmail,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	token.Header["typ"] = identityTokenType
	signed, err := token.SignedString([]byte(identitySecret))
	require.NoError(t, err, "sign identity token")
	return signed
}
