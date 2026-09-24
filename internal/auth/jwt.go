package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenUseClaim   = "token_use"
	accessTokenUse  = "access"
	refreshTokenUse = "refresh"
)

var ErrInvalidToken = errors.New("auth: invalid or expired token")

type Config struct {
	Issuer             string
	Audience           string
	SecretKey          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type claims struct {
	Username string `json:"unique_name"`
	TokenUse string `json:"token_use"`
	jwt.RegisteredClaims
}

type JWTService struct {
	config     *Config
	signingKey []byte
}

func NewJWTService(config *Config) *JWTService {
	return &JWTService{config: config, signingKey: []byte(config.SecretKey)}
}

func (s *JWTService) GenerateAccessToken(userId int, username string) (string, error) {
	return s.generateToken(userId, username, accessTokenUse, s.config.AccessTokenExpiry)
}

func (s *JWTService) GenerateRefreshToken(userId int, username string) (string, error) {
	return s.generateToken(userId, username, refreshTokenUse, s.config.RefreshTokenExpiry)
}

func (s *JWTService) ValidateAccessToken(token string) (userId int, username string, err error) {
	return s.validateToken(token, accessTokenUse)
}

func (s *JWTService) ValidateRefreshToken(token string) (userId int, username string, err error) {
	return s.validateToken(token, refreshTokenUse)
}

func (s *JWTService) generateToken(userId int, username, tokenUse string, expiry time.Duration) (string, error) {
	jti, err := randomJti()
	if err != nil {
		return "", err
	}

	now := time.Now()
	c := claims{
		Username: username,
		TokenUse: tokenUse,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userId),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			Issuer:    s.config.Issuer,
			Audience:  jwt.ClaimStrings{s.config.Audience},
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.signingKey)
}

func (s *JWTService) validateToken(tokenStr, expectedUse string) (int, string, error) {
	var c claims

	_, err := jwt.ParseWithClaims(tokenStr, &c, func(*jwt.Token) (any, error) {
		return s.signingKey, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.config.Issuer),
		jwt.WithAudience(s.config.Audience),
	)
	if err != nil {
		return 0, "", ErrInvalidToken
	}

	if c.TokenUse != expectedUse {
		return 0, "", ErrInvalidToken
	}

	userId, err := strconv.Atoi(c.Subject)
	if err != nil {
		return 0, "", ErrInvalidToken
	}

	return userId, c.Username, nil
}

func randomJti() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
