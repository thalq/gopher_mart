package models

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	jwt.StandardClaims
	UserID int64 `json:"user_id"`
}

type Order struct {
	OrderID    string    `json:"order_id"`
	Status     string    `json:"status"`
	UploadTime time.Time `json:"upload_time"`
}

type Balance struct {
	Current   int64 `json:"current"`
	Withdrawn int64 `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string `json:"order"`
	Sum   int64  `json:"sum"`
}

type WithdrawResponse struct {
	OrderID     string    `json:"order"`
	Sum         int64     `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type AccrualInfo struct {
	OrderID string `json:"order"`
	Status  string `json:"status"`
	Accrual int64  `json:"accrual"`
}
