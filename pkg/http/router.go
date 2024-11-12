package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thalq/gopher_mart/internal/auth"
	"github.com/thalq/gopher_mart/internal/constants"
	myMiddleware "github.com/thalq/gopher_mart/internal/middleware"
	"github.com/thalq/gopher_mart/internal/orders"
	"github.com/thalq/gopher_mart/pkg/config"
	"github.com/thalq/gopher_mart/pkg/storage"
)

func NewRouter(cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(myMiddleware.Logging)
	r.Use(myMiddleware.AuthMiddleware(constants.JWTSecret))

	db := storage.GetDB()
	authService := auth.NewAuthService(db, constants.JWTSecret)
	authHandler := auth.NewAuthHandler(authService)
	orderService := orders.NewOrderService(db)
	orderHandler := orders.NewOrderHandler(orderService)
	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/login", authHandler.Login)
	r.Post("/api/user/orders", orderHandler.UploadOrder)
	return r
}
