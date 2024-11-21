package orders

import (
	"database/sql"

	"net/http"

	logger "github.com/thalq/gopher_mart/internal/middleware"
	"github.com/thalq/gopher_mart/internal/models"
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

func (s *OrderService) CreateOrder(
	userID int64,
	orderNumber string,
	accrualInfo models.AccrualInfo,
) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		"INSERT INTO orders (user_id, order_id, status) VALUES ($1, $2, $3, $4)",
		userID,
		orderNumber,
		accrualInfo.Status,
		accrualInfo.Accrual,
	)
	if err != nil {
		logger.Sugar.Errorf("Failed to insert order: %v", err)
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO user_balance (user_id, current_balance)
		VALUES ($1, 0.0)
		ON CONFLICT (user_id) DO NOTHING
	`, userID)
	if err != nil {
		logger.Sugar.Errorf("Failed to insert user balance: %v", err)
		return err
	}
	_, err = tx.Exec(
		"UPDATE user_balance SET current_balance = current_balance + $1 WHERE user_id = $2",
		accrualInfo.Accrual,
		userID,
	)
	if err != nil {
		logger.Sugar.Errorf("Failed to update user balance: %v", err)
		return err
	}
	if err = tx.Commit(); err != nil {
		logger.Sugar.Errorf("Failed to commit transaction: %v", err)
		return err
	}
	logger.Sugar.Infof("Order %s created for user %s", orderNumber, userID)
	return nil
}

func (s *OrderService) CheckOtherUserHasOrders(orderNumber string) (bool, error) {
	var orderExists bool
	if err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM orders WHERE order_id = $1)", orderNumber).Scan(&orderExists); err != nil {
		return false, err
	}
	return orderExists, nil
}

func (s *OrderService) GetOrders(userID int64) ([]models.Order, error) {
	rows, err := s.db.Query(
		"SELECT order_id, status, upload_time FROM orders WHERE user_id = $1 ORDER BY upload_time DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.OrderID, &order.Status, &order.UploadTime); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		logger.Sugar.Errorf("Failed to iterate over rows: %v", err)
		return nil, err
	}
	logger.Sugar.Infof("Got orders for user %s: %v", userID, orders)

	return orders, nil
}

func (s *OrderService) GetBalance(userID int64) (models.Balance, error) {
	var balance models.Balance
	if err := s.db.QueryRow("SELECT SUM(withdrawal), SUM(current) FROM orders WHERE user_id = $1", userID).Scan(&balance.Withdrawn, &balance.Current); err != nil {
		return balance, err
	}
	logger.Sugar.Infof("Got balance for user %s: %v", userID, balance)

	return balance, nil
}

func (s *OrderService) WithdrawRequest(
	userID int64,
	orderId string,
	sum float32,
	accrualInfo models.AccrualInfo,
) int {
	tx, err := s.db.Begin()
	if err != nil {
		return http.StatusBadRequest
	}
	defer tx.Rollback()

	var balance float32
	if err := tx.QueryRow("SELECT current_balance FROM user_balance WHERE user_id = $1", userID).Scan(&balance); err != nil {
		logger.Sugar.Errorf("Failed to get balance for user %d: %v", userID, err)
		return http.StatusInternalServerError
	}
	balance += accrualInfo.Accrual
	if balance < sum {
		logger.Sugar.Errorf("Not enough money for user %d", userID)
		return http.StatusPaymentRequired
	}

	_, err = tx.Exec(
		"INSERT INTO orders (user_id, order_id, withdrawal, status, accrual) VALUES ($1, $2, $3, $4, $5)",
		userID,
		orderId,
		sum,
		accrualInfo.Status,
		accrualInfo.Accrual,
	)
	if err != nil {
		logger.Sugar.Errorf("Failed to insert order: %v", err)
		return http.StatusInternalServerError
	}

	_, err = tx.Exec(`
		INSERT INTO user_balance (user_id, current_balance)
		VALUES ($1, 0.0)
		ON CONFLICT (user_id) DO NOTHING
	`, userID)
	if err != nil {
		logger.Sugar.Errorf("Failed to insert user balance: %v", err)
		return http.StatusInternalServerError
	}
	_, err = tx.Exec(
		"UPDATE user_balance SET current_balance = current_balance - $1 WHERE user_id = $2",
		sum,
		userID,
	)
	if err != nil {
		logger.Sugar.Errorf("Failed to update user balance: %v", err)
		return http.StatusInternalServerError
	}

	if err = tx.Commit(); err != nil {
		logger.Sugar.Errorf("Failed to commit transaction: %v", err)
		return http.StatusInternalServerError
	}

	logger.Sugar.Infof("Withdraw %d for user %d", sum, userID)
	return http.StatusOK
}

func (s *OrderService) GetUserWithdrawls(userID int64) ([]models.WithdrawResponse, error) {
	rows, err := s.db.Query(
		"SELECT order_id, withdrawal, upload_time FROM orders WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawls []models.WithdrawResponse
	for rows.Next() {
		var withdrawl models.WithdrawResponse
		if err := rows.Scan(&withdrawl.OrderID, &withdrawl.Sum, &withdrawl.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawls = append(withdrawls, withdrawl)
	}
	if err = rows.Err(); err != nil {
		logger.Sugar.Errorf("Failed to iterate over rows: %v", err)
		return nil, err
	}
	logger.Sugar.Infof("Got withdrawls for user %d: %v", userID, withdrawls)

	return withdrawls, nil
}
