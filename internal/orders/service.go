package orders

import (
	"database/sql"

	logger "github.com/thalq/gopher_mart/internal/middleware"
)

type OrderService struct {
	db *sql.DB
}

func NewOrderService(db *sql.DB) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) CheckUserHasOrders(userID int64, orderNumber string) (bool, error) {
	var orderExists bool
	if err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM orders WHERE user_id = $1 AND order_id = $2)", userID, orderNumber).Scan(&orderExists); err != nil {
		return false, err
	}
	return orderExists, nil
}

func (s *OrderService) CreateOrder(userID int64, orderNumber string) error {
	_, err := s.db.Exec("INSERT INTO orders (user_id, order_id) VALUES ($1, $2)", userID, orderNumber)
	logger.Sugar.Infof("Insert order %s for user %s", orderNumber, userID)
	return err
}

func (s *OrderService) CheckOtherUserHasOrders(orderNumber string) (bool, error) {
	var orderExists bool
	if err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM orders WHERE order_id = $1)", orderNumber).Scan(&orderExists); err != nil {
		return false, err
	}
	return orderExists, nil
}
