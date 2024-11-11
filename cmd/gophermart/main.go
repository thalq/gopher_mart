package main

import (
	"net/http"

	logger "github.com/thalq/gopher_mart/internal/middleware"
	"github.com/thalq/gopher_mart/pkg/config"
	router "github.com/thalq/gopher_mart/pkg/http"
	"github.com/thalq/gopher_mart/pkg/storage"
)

func main() {
	logger.InitLogger()
	cfg := config.NewConfig()

	storage.InitDB(cfg.DatabaseURI)
	router := router.NewRouter(cfg)

	logger.Sugar.Infof("Starting server on %s", cfg.RunAdress)
	if err := http.ListenAndServe(cfg.RunAdress, router); err != nil {
		logger.Sugar.Fatalf("Error run server: %s", err)
	}

}
