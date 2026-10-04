package auth

import (
	"acrocuit/internal/options"
	"acrocuit/internal/response"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	IdentityHeader          = "X-Forwarded-Identity"
	UserEmailHeader         = "X-Forwarded-User-Email"
	MinIdentitySecretLength = 32
	SingleUserEmail         = "owner@localhost"

	identityTokenType   = "identity+jwt"
	contextUserEmailKey = "userEmail"
)

var errInvalidIdentity = errors.New("auth: invalid identity token")

func RequireIdentity(opts *options.Options) (gin.HandlerFunc, error) {
	switch opts.Auth.Mode {
	case options.AuthModeSingleUser:
		return requireSingleUser(), nil
	case options.AuthModeForwardedHeader:
		return requireForwardedHeader(), nil
	case options.AuthModeIdentityToken:
		return requireIdentityToken(opts.Auth.IdentitySecret)
	default:
		return nil, fmt.Errorf("auth: unsupported auth mode %q", opts.Auth.Mode)
	}
}

func requireSingleUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(contextUserEmailKey, SingleUserEmail)
		c.Next()
	}
}

func requireForwardedHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		email := strings.TrimSpace(c.GetHeader(UserEmailHeader))
		if email == "" {
			response.AbortError(c, http.StatusUnauthorized, "Missing or invalid identity.")
			return
		}

		c.Set(contextUserEmailKey, email)
		c.Next()
	}
}

func requireIdentityToken(identitySecret string) (gin.HandlerFunc, error) {
	secret := []byte(identitySecret)
	if len(secret) < MinIdentitySecretLength {
		return nil, fmt.Errorf("auth: IDENTITY_JWT_SECRET must be at least %d bytes", MinIdentitySecretLength)
	}

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)

	return func(c *gin.Context) {
		email, err := parseIdentity(parser, secret, c.GetHeader(IdentityHeader))
		if err != nil {
			response.AbortError(c, http.StatusUnauthorized, "Missing or invalid identity.")
			return
		}

		c.Set(contextUserEmailKey, email)
		c.Next()
	}, nil
}

func parseIdentity(parser *jwt.Parser, secret []byte, token string) (string, error) {
	if token == "" {
		return "", errInvalidIdentity
	}

	var claims jwt.RegisteredClaims
	_, err := parser.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if t.Header["typ"] != identityTokenType {
			return nil, errInvalidIdentity
		}
		return secret, nil
	})
	if err != nil {
		return "", err
	}

	if claims.Subject == "" {
		return "", errInvalidIdentity
	}

	return claims.Subject, nil
}

func UserEmail(c *gin.Context) string {
	email, _ := c.Get(contextUserEmailKey)
	s, _ := email.(string)
	return s
}
