package models

import "github.com/golang-jwt/jwt"

type Claims struct {
	jwt.StandardClaims
	UserID int64 `json:"user_id"`
}
