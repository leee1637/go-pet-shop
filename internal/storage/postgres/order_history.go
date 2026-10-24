package postgres

import (
	"context"
	"fmt"
	"go-pet-shop/internal/models"
)

// GetUserOrderHistory(email string) ([]models.OrderDetail, error)
// GetPopularProducts() ([]models.PopularProduct, error)

// type OrderDetail struct {
//     OrderID           int
//     CreatedAt         time.Time
//     TotalPrice        string
//     TransactionStatus string
//     Items             []OrderItemDetail
// }

// type OrderItemDetail struct {
//     ProductID   int
//     ProductName string
//     Quantity    int
//     UnitPrice   string
// }

func (s *Storage) GetUserOrderHistory(ctx context.Context, email string) ([]models.OrderDetail, error) {
	query := `SELECT o.id, o.created_at, o.total_price, t.status AS tx_status,
	oi.product_id, p.name, p.stock, p.price AS unit_price FROM orders o
	JOIN users u ON o.user_id = u.id
	LEFT JOIN transactions t ON t.order_id = o.id
	JOIN order_items oi ON oi.order_id = o.id
	JOIN products p ON oi.product_id = p.id
	WHERE email = $1 ORDER BY o.created_at DESC, o.id DESC, oi.id;`

	row, err := s.db.Query(ctx, query, email)
	if err != nil {
		return []models.OrderDetail{}, fmt.Errorf("error: %w", err)
	}

	orderDetail := []models.OrderDetail{}

	for row.Next() {
		item := models.OrderDetail{}
		err := row.Scan(item.OrderID, item.CreatedAt, item.TotalPrice, item.TransactionStatus, item.)
	}

}
