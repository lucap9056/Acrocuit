package handlers

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/models"
	"acrocuit/internal/response"
	"acrocuit/internal/storage"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/pbkdf2"
)

const (
	pbkdf2SaltLength = 32
	pbkdf2Iterations = 100000
	pbkdf2KeyLength  = 32
)

type authController struct {
	*storage.Storage
	jwt *auth.JWTService
}

func authHandlers(r *gin.RouterGroup, s *storage.Storage, jwtService *auth.JWTService) {
	c := &authController{s, jwtService}

	r.POST("register", c.Register)
	r.POST("login", c.Login)
	r.POST("refresh", c.Refresh)
}

type userCredentials struct {
	Username string `json:"username" binding:"required,min=1,max=127"`
	Password string `json:"password" binding:"required"`
}

type tokenPair struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (ctr *authController) Register(c *gin.Context) {
	var req userCredentials
	if !bindJSON(c, &req) {
		return
	}

	salt := make([]byte, pbkdf2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		internalError(c, err)
		return
	}
	hash := hashPassword(req.Password, salt)

	_, err := ctr.Users.CreateUser(c.Request.Context(), req.Username, salt, hash)
	if errors.Is(err, storage.ErrConflict) {
		response.Error(c, http.StatusConflict, "Username is already taken. Please choose another one.")
		return
	}
	if err != nil {
		internalError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "Registration successful."})
}

func (ctr *authController) Login(c *gin.Context) {
	var req userCredentials
	if !bindJSON(c, &req) {
		return
	}

	secret, err := ctr.Users.GetUserSecret(c.Request.Context(), req.Username)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		internalError(c, err)
		return
	}
	if err != nil || !verifyPassword(req.Password, secret) {
		response.Error(c, http.StatusUnauthorized, "Invalid username or password.")
		return
	}

	accessToken, err := ctr.jwt.GenerateAccessToken(secret.UserId, req.Username)
	if err != nil {
		internalError(c, err)
		return
	}
	refreshToken, err := ctr.jwt.GenerateRefreshToken(secret.UserId, req.Username)
	if err != nil {
		internalError(c, err)
		return
	}

	response.OK(c, tokenPair{AccessToken: accessToken, RefreshToken: refreshToken})
}

func (ctr *authController) Refresh(c *gin.Context) {
	var req tokenPair
	if !bindJSON(c, &req) {
		return
	}

	userId, username, err := ctr.jwt.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid refresh token.")
		return
	}

	accessToken, err := ctr.jwt.GenerateAccessToken(userId, username)
	if err != nil {
		internalError(c, err)
		return
	}
	refreshToken, err := ctr.jwt.GenerateRefreshToken(userId, username)
	if err != nil {
		internalError(c, err)
		return
	}

	response.OK(c, tokenPair{AccessToken: accessToken, RefreshToken: refreshToken})
}

func hashPassword(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, pbkdf2KeyLength, sha256.New)
}

func verifyPassword(pswd string, secret *models.UserSecret) bool {
	hash := hashPassword(pswd, secret.PasswordSalt)
	return subtle.ConstantTimeCompare(hash, secret.PasswordHash) == 1
}
