package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	logger "github.com/thalq/gopher_mart/internal/middleware"
	"github.com/thalq/gopher_mart/pkg/storage"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (req *RegisterRequest) Validate() error {
	if req.Login == "" {
		return fmt.Errorf("login is empty")
	}
	if req.Password == "" {
		return fmt.Errorf("password is empty")
	}
	return nil
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	logger.Sugar.Infof("Got request: %s", string(body))
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Не удалось распарсить JSON", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	db := storage.GetDB()
	logger.Sugar.Infoln(db)
	var userExists bool
	if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", req.Login).Scan(&userExists); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if userExists {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	if _, err := db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", req.Login, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := h.service.GenerateToken(req.Login)
	http.SetCookie(w, &http.Cookie{
		Name:    "Authorization",
		Value:   token,
		Expires: time.Now().Add(time.Hour * 24),
		Path:    "/",
	})
	w.WriteHeader(http.StatusOK)
}
