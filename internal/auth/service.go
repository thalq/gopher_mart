package auth

import (
	"errors"
	"time"

	"database/sql"

	"github.com/golang-jwt/jwt"
	logger "github.com/thalq/gopher_mart/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db        *sql.DB
	jwtSecret string
}

func NewAuthService(db *sql.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: jwtSecret}
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

func (s *AuthService) CheckUserExists(username string) (bool, error) {
	var userExists bool
	if err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&userExists); err != nil {
		logger.Sugar.Errorf("Error check user exists: %s", err)
		return false, err
	}
	return userExists, nil
}

func (s *AuthService) Register(login, password string) error {
	hash, err := s.HashPassword(password)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", login, hash)
	if err != nil {
		logger.Sugar.Errorf("Error insert user to db: %s", err)
		return err
	}
	return nil
}

func (s *AuthService) Authenticate(login, password string) (bool, error) {
	var storedPassword string
	err := s.db.QueryRow("SELECT password FROM users WHERE username = $1", login).Scan(&storedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Sugar.Infof("User %s not found", login)
			return false, nil
		}
		logger.Sugar.Errorf("Error get user from db: %s", err)
		return false, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)); err != nil {
		logger.Sugar.Infof("Password for user %s is incorrect", login)
		return false, nil
	}
	return true, nil
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
