package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thalq/gopher_mart/internal/constants"
	logger "github.com/thalq/gopher_mart/internal/middleware"
	"github.com/thalq/gopher_mart/pkg/config"
	router "github.com/thalq/gopher_mart/pkg/http"
	"github.com/thalq/gopher_mart/pkg/storage"
)

func TestHandlers(t *testing.T) {
	logger.InitLogger()
	cfg := config.NewConfig()
	storage.InitDB(cfg.DatabaseURI)
	r := router.NewRouter(cfg)

	t.Run("Login", func(t *testing.T) {
		reqBody := `{"login": "testUser", "password": "test"}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Orders", func(t *testing.T) {

		authCookie := &http.Cookie{
			Name:    "Authorization",
			Value:   constants.TestToken,
			Expires: time.Now().Add(time.Hour * 24),
			Path:    "/",
		}
		reqBody := `12345678903`
		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(authCookie)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

	})

	t.Run("GetOrders", func(t *testing.T) {
		authCookie := &http.Cookie{
			Name:    "Authorization",
			Value:   constants.TestToken,
			Expires: time.Now().Add(time.Hour * 24),
			Path:    "/",
		}
		req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		req.AddCookie(authCookie)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		expectedBody := `[{"order_id":"","status":"NEW","upload_time":"2024-11-15T10:44:29.359919Z"},{"order_id":"12345678903","status":"NEW","upload_time":"2024-11-15T10:41:22.81692Z"}]`
		assert.Equal(t, expectedBody, w.Body.String())
	})

	t.Run("GetBalance", func(t *testing.T) {
		authCookie := &http.Cookie{
			Name:    "Authorization",
			Value:   constants.TestToken,
			Expires: time.Now().Add(time.Hour * 24),
			Path:    "/",
		}
		req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
		req.AddCookie(authCookie)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		expectedBody := `{"current":0,"withdrawn":0}`
		assert.Equal(t, expectedBody, w.Body.String())
	})

}
