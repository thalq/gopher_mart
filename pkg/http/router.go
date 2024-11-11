package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thalq/gopher_mart/internal/auth"
	myMiddleware "github.com/thalq/gopher_mart/internal/middleware"
	"github.com/thalq/gopher_mart/pkg/config"
	"github.com/thalq/gopher_mart/pkg/storage"
)

func NewRouter(cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(myMiddleware.Logging)

	db := storage.GetDB()
	authService := auth.NewAuthService(db, "supersecretkey")
	authHandler := auth.NewAuthHandler(authService)
	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/login", authHandler.Login)
	return r
}
