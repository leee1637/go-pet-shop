package order

import (
	"context"
	"go-pet-shop/internal/models"
)

type OrdersMock struct {
	CreateOrderFunc            func(ctx context.Context, order models.Order) (int, error)
	AddOrderItemFunc           func(ctx context.Context, orderItem models.OrderItem) error
	GetOrderByIDFunc           func(ctx context.Context, id int) (models.Order, error)
	GetOrdersByUserEmailFunc   func(ctx context.Context, email string) ([]models.Order, error)
	GetOrderItemsByOrderIDFunc func(ctx context.Context, orderID int) ([]models.OrderItem, error)
	GetUserOrderHistoryFunc    func(ctx context.Context, email string) ([]models.OrderDetail, error)
	PlaceOrderFunc             func(ctx context.Context, userEmail string, items []models.OrderItem) (int, error)
}

func (m *OrdersMock) CreateOrder(ctx context.Context, order models.Order) (int, error) {
	if m.CreateOrderFunc != nil {
		return m.CreateOrderFunc(ctx, order)
	}
	return 0, nil
}

func (m *OrdersMock) AddOrderItem(ctx context.Context, orderItem models.OrderItem) error {
	if m.AddOrderItemFunc != nil {
		return m.AddOrderItemFunc(ctx, orderItem)
	}
	return nil
}

func (m *OrdersMock) GetOrderByID(ctx context.Context, id int) (models.Order, error) {
	if m.GetOrderByIDFunc != nil {
		return m.GetOrderByIDFunc(ctx, id)
	}
	return models.Order{}, nil
}

func (m *OrdersMock) GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error) {
	if m.GetOrdersByUserEmailFunc != nil {
		return m.GetOrdersByUserEmailFunc(ctx, email)
	}
	return []models.Order{}, nil
}

func (m *OrdersMock) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	if m.GetOrderItemsByOrderIDFunc != nil {
		return m.GetOrderItemsByOrderIDFunc(ctx, orderID)
	}
	return []models.OrderItem{}, nil
}

func (m *OrdersMock) GetUserOrderHistory(ctx context.Context, email string) ([]models.OrderDetail, error) {
	if m.GetUserOrderHistoryFunc != nil {
		return m.GetUserOrderHistoryFunc(ctx, email)
	}
	return []models.OrderDetail{}, nil
}

func (m *OrdersMock) PlaceOrder(ctx context.Context, userEmail string, items []models.OrderItem) (int, error) {
	if m.PlaceOrderFunc != nil {
		return m.PlaceOrderFunc(ctx, userEmail, items)
	}
	return 0, nil
}
