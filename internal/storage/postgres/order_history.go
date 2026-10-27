package postgres

import (
	"context"
	"fmt"
	"go-pet-shop/internal/models"
	"time"
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
	query := `SELECT o.id, o.created_at, o.total_price, COALESCE(t.status, 'unknown') AS tx_status ,
	oi.product_id, p.name, oi.quantity, p.price AS unit_price FROM orders o
	JOIN users u ON o.user_id = u.id
	LEFT JOIN transactions t ON t.order_id = o.id
	JOIN order_items oi ON oi.order_id = o.id
	JOIN products p ON oi.product_id = p.id
	WHERE email = $1 ORDER BY o.created_at DESC, o.id DESC, oi.id;`

	row, err := s.db.Query(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	defer row.Close()

	orderDetail := []models.OrderDetail{}
	var current *models.OrderDetail

	for row.Next() {
		var (
			orderID    int
			createdAt  time.Time
			totalPrice string
			txStatus   string
			item       models.OrderItemDetail
		)

		err := row.Scan(
			&orderID, &createdAt, &totalPrice, &txStatus,
			&item.ProductID, &item.ProductName, &item.Quantity, &item.UnitPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("insert scan: %w", err)
		}

		if current == nil || current.OrderID != orderID {
			orderDetail = append(orderDetail, models.OrderDetail{
				OrderID:           orderID,
				CreatedAt:         createdAt,
				TotalPrice:        totalPrice,
				TransactionStatus: txStatus,
			})
			current = &orderDetail[len(orderDetail)-1]
		}
		current.Items = append(current.Items, item)
	}

	return orderDetail, nil
}

func (s *Storage) GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error) {
	query := `SELECT p.id, p.name, SUM(oi.quantity) AS total_sold FROM products 
	JOIN order_items oi ON p.id = oi.product_id GROUP BY p.id, p.name ORDER BY total_sold DESC`

	row, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	defer row.Close()

	pp := []models.PopularProduct{}

	for row.Next() {
		var (
			id   int
			name string
			sum  int
		)
		err := row.Scan(&id, &name, &sum)
		if err != nil {
			return nil, fmt.Errorf("error: %w", err)
		}
		pp = append(pp, models.PopularProduct{
			ProductID:   id,
			ProductName: name,
			TotalSold:   sum,
		})
	}

	return pp, nil
}
