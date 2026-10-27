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

// Версия 3: оформление заказа с транзакцией
// Цели:

// Реализовать полноценное оформление заказа в транзакции.
// Обновить остатки товара (stock).
// Создать заказ, позиции заказа и транзакцию оплаты атомарно.
// Методы (storage interface):

func (s *Storage) PlaceOrder(ctx context.Context, userEmail string, items []models.OrderItem) (int, error) {
	tx, err := s.db.Begin(ctx)

	if err != nil {
		return 0, fmt.Errorf("error start transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	userID := 0
	err = tx.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, userEmail).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("get user by email: %w", err)
	}

	allPrice := 0

	for _, item := range items {
		price := 0
		err := tx.QueryRow(ctx, `UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1
		RETURNING price`, item.Quantity, item.ProductID).Scan(&price)
		allPrice += price * item.Quantity
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return 0, fmt.Errorf("product %d: %w", item.ProductID, ErrNotFound)
			}
			return 0, fmt.Errorf("decrease stock product %d: %w", item.ProductID, err)
		}

		allPrice = allPrice + price
	}

	var orderID int
	err = tx.QueryRow(ctx, `INSERT INTO orders (user_id, total_price) VALUES ($1, $2) RETURNING id`, userID, allPrice).Scan(&orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("get user by email: %w", err)
	}

	for _, item := range items {
		_, err = tx.Query(ctx, `INSERT INTO order_items (order_id, product_id, quantity)
		 VALUES ($1, $2, $3)`, orderID, item.ProductID, item.Quantity)
		if err != nil {
			return 0, fmt.Errorf("decrease stock product %d: %w", item.ProductID, err)
		}
	}

	_, err = tx.Exec(ctx, `INSERT INTO transactions (order_id, amount, status) 
	VALUES ($1, $2, $3)`, orderID, allPrice, "created")
	if err != nil {
		return 0, fmt.Errorf("insert transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("insert commit: %w", err)
	}

	return orderID, nil
}

// Внутри PlaceOrder:
// Начать транзакцию
// Проверить и уменьшить stock товаров (UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1; 0 затронутых строк значит, что товара не хватает, и заказ откатывается)
// Создать заказ (INSERT INTO orders)
// Добавить позиции заказа (INSERT INTO order_items)
// Создать запись транзакции (INSERT INTO transactions)
// Commit или Rollback в случае ошибки
