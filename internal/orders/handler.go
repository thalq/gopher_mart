package orders

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/thalq/gopher_mart/internal/constants"
	logger "github.com/thalq/gopher_mart/internal/middleware"
	"github.com/thalq/gopher_mart/internal/models"
)

type OrderHandler struct {
	service *OrderService
}

func NewOrderHandler(service *OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, ok := ctx.Value(constants.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	orderNumber := strings.TrimSpace(string(body))

	if !ValidateOrderNumber(orderNumber) {
		http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}
	userHasOrder, err := h.service.CheckUserHasOrders(userID, orderNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if userHasOrder {
		w.WriteHeader(http.StatusOK)
		logger.Sugar.Infof("User %s has order %s", userID, orderNumber)
	} else {
		otherUserHasOrder, err := h.service.CheckOtherUserHasOrders(orderNumber)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if otherUserHasOrder {
			http.Error(w, "Order already exists", http.StatusConflict)
			logger.Sugar.Infof("Order %s already exists for another user", orderNumber)
		} else {
			if err := h.service.CreateOrder(userID, orderNumber); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusAccepted)
			logger.Sugar.Infof("User %s created order %s", userID, orderNumber)
		}
	}

}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, ok := ctx.Value(constants.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.service.GetOrders(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		http.Error(w, "No orders for user", http.StatusNoContent)
		return
	}
	logger.Sugar.Infof("Got %d orders for user", len(orders))
	response, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(response)
	w.WriteHeader(http.StatusOK)
}

func (h *OrderHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, ok := ctx.Value(constants.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.service.GetBalance(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Sugar.Infof("Got balance for user %s", userID)
	response, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(response)
	w.WriteHeader(http.StatusOK)
}

func (h *OrderHandler) WithdrawRequest(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, ok := ctx.Value(constants.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}
	logger.Sugar.Infof("User %s requested withdraw", userID)

	var request models.WithdrawRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Sugar.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if err := json.Unmarshal(body, &request); err != nil {
		logger.Sugar.Errorf("Failed to unmarshal request: %v", err)
		http.Error(w, "Failed to unmarshal request", http.StatusBadRequest)
		return
	}
	logger.Sugar.Infof("Got withdraw request: %v", request)

	if !ValidateOrderNumber(request.Order) {
		http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}

	response := h.service.WithdrawRequest(userID, request.Order, request.Sum)
	w.WriteHeader(response)
}

func (h *OrderHandler) UserWithdrawls(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, ok := ctx.Value(constants.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawls, err := h.service.GetUserWithdrawls(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(withdrawls) == 0 {
		http.Error(w, "No withdrawls for user", http.StatusNoContent)
		return
	}
	logger.Sugar.Infof("Got %d withdrawls for user", len(withdrawls))
	response, err := json.Marshal(withdrawls)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(response)
	w.WriteHeader(http.StatusOK)
}

func (h *OrderHandler) OrderAccrual(w http.ResponseWriter, r *http.Request) {
	orderNumber := chi.URLParam(r, "number")
	if orderNumber == "" {
		http.Error(w, "Order number is required", http.StatusBadRequest)
		return
	}
	logger.Sugar.Infof("Got request for order %s", orderNumber)

	order, err := h.service.OrderAccrual(orderNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if order.OrderID == "" {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	logger.Sugar.Infof("Got order %s", orderNumber)
	response, err := json.Marshal(order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(response)
	w.WriteHeader(http.StatusOK)
}
