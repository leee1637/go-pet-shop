package postgres

import (
	"context"
	"errors"
	"fmt"
	"go-pet-shop/internal/models"

	"github.com/jackc/pgx/v5"
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
	query := `INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)`

	_, err := s.db.Query(ctx, query, orderItem.OrderID, orderItem.ProductID, orderItem.Quantity)
	if err != nil {
		return fmt.Errorf("create order item: %w", err)
	}

	return nil
}

func (s *Storage) GetOrderByID(ctx context.Context, id int) (models.Order, error) {
	query := `SELECT id, user_id, total_price, created_at FROM orders WHERE id = $1`

	m := models.Order{}

	err := s.db.QueryRow(ctx, query, id).Scan(&m.ID, &m.UserID, &m.TotalPrice, &m.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, ErrNotFound
		}

		return models.Order{}, fmt.Errorf("error: %w", err)
	}

	return m, nil
}

func (s *Storage) GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error) {
	query := `SELECT o.id, o.user_id, o.total_price, o.created_at FROM orders o 
	JOIN users u ON o.user_id = u.id
	WHERE u.email = $1`

	row, err := s.db.Query(ctx, query, email)

	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	sl := []models.Order{}

	for row.Next() {
		o := models.Order{}
		err := row.Scan(&o.ID, &o.UserID, &o.TotalPrice, &o.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error: %w", err)
		}
		sl = append(sl, o)
	}

	return sl, nil
}

func (s *Storage) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	query := `SELECT i.id, i.order_id, i.product_id, i.quantity FROM order_items i
	JOIN orders o i ON i.order_id = o.id
	WHERE o.id = $1`

	row, err := s.db.Query(ctx, query, orderID)

	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	sl := []models.OrderItem{}

	for row.Next() {
		o := models.OrderItem{}
		err := row.Scan(&o.ID, &o.OrderID, &o.ProductID, &o.Quantity)
		if err != nil {
			return nil, fmt.Errorf("error: %w", err)
		}
		sl = append(sl, o)
	}

	return sl, nil
}
