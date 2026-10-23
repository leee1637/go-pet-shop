package postgres

import (
	"context"
	"fmt"
	"go-pet-shop/internal/models"
)

// CreateOrder(order models.Order) (int, error) // Возвращает ID созданного заказа
// AddOrderItem(orderItem models.OrderItem) error
// GetOrderByID(id int) (models.Order, error)
// GetOrdersByUserEmail(email string) ([]models.Order, error)
// GetOrderItemsByOrderID(orderID int) ([]models.OrderItem, error)

func (s *Storage) CreateOrder(ctx context.Context, order models.Order) (int, error) {
	query := `INSERT INTO orders (user_id, total_price) VALUES ($1, $2)
	RETURNING id`

	var id int

	err := s.db.QueryRow(ctx, query, order.UserID, order.TotalPrice).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create order: %w", err)
	}

	return id, nil

}

func (s *Storage) AddOrderItem(ctx context.Context, orderItem models.OrderItem) error {
	query := `INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)
	RETURNING id`

	row, err := s.db.Query(ctx, query, orderItem.OrderID, orderItem.ProductID, orderItem.Quantity)
	if err != nil {
		return fmt.Errorf("create order item: %w", err)
	}

	if row == 0 {

	}

	return nil
}

func (s *Storage) GetOrderByID(ctx context.Context, id int) (models.Order, error) {

}

func (s *Storage) GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error) {

}

func (s *Storage) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.Order, error) {

}
