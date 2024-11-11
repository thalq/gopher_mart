package auth

import (
	"time"

	"github.com/golang-jwt/jwt"
	logger "github.com/thalq/gopher_mart/internal/middleware"
)

type AuthService struct {
	jwtSecret string
}

func NewAuthService(jwtSecret string) *AuthService {
	return &AuthService{jwtSecret: jwtSecret}
}

func (s *AuthService) GenerateToken(username string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   username,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		logger.Sugar.Errorf("Error generate token: %s", err)
	}
	return tokenString
}
