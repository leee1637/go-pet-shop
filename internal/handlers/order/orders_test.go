package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Роут как в main.go (с теми же URL-паттернами).
func newTestRouter(mock *OrdersMock) *gin.Engine {
	h := New(testLogger(), mock)

	r := gin.New()
	r.POST("/orders", h.CreateOrder)
	r.GET("/orders/:id", h.GetOrderByID)
	r.GET("/orders/:id/items", h.GetOrderItemsByOrderID)
	r.GET("/users/:email/orders", h.GetOrdersByUserEmail)
	r.GET("/users/:email/history", h.GetUserOrderHistory)
	r.POST("/checkout", h.PlaceOrder)
	return r
}

func performRequest(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

// =======================
// CreateOrder
// =======================

func TestCreateOrder_Success(t *testing.T) {
	mock := &OrdersMock{
		CreateOrderFunc: func(ctx context.Context, o models.Order) (int, error) {
			if o.UserID != 1 || o.TotalPrice != "100.00" {
				return 0, fmt.Errorf("unexpected order: %+v", o)
			}
			return 42, nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPost, "/orders", `{"user_id":1,"total_price":"100.00"}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("want code %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if id, ok := resp["id"].(float64); !ok || int(id) != 42 {
		t.Fatalf("want id 42 in response, got %s", w.Body.String())
	}
}

func TestCreateOrder_BadRequest(t *testing.T) {
	r := newTestRouter(&OrdersMock{})
	w := performRequest(r, http.MethodPost, "/orders", `{"user_id":`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateOrder_Fail(t *testing.T) {
	mock := &OrdersMock{
		CreateOrderFunc: func(ctx context.Context, o models.Order) (int, error) {
			return 0, errors.New("db down")
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPost, "/orders", `{"user_id":1,"total_price":"100.00"}`)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =======================
// GetOrderByID
// =======================

func TestGetOrderByID_Success(t *testing.T) {
	mock := &OrdersMock{
		GetOrderByIDFunc: func(ctx context.Context, id int) (models.Order, error) {
			if id != 1 {
				return models.Order{}, fmt.Errorf("unexpected id: %d", id)
			}
			return models.Order{ID: 1, UserID: 2, TotalPrice: "100.00", CreatedAt: time.Now()}, nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/orders/1", "")

	if w.Code != http.StatusOK {
		t.Fatalf("want code %d, got %d", http.StatusOK, w.Code)
	}

	var got models.Order
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.ID != 1 || got.UserID != 2 {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestGetOrderByID_NotFound(t *testing.T) {
	mock := &OrdersMock{
		GetOrderByIDFunc: func(ctx context.Context, id int) (models.Order, error) {
			return models.Order{}, storage.ErrNotFound
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/orders/999", "")

	if w.Code != http.StatusNotFound {
		t.Fatalf("want code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetOrderByID_BadID(t *testing.T) {
	r := newTestRouter(&OrdersMock{})
	w := performRequest(r, http.MethodGet, "/orders/not-a-number", "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetOrderByID_Fail(t *testing.T) {
	mock := &OrdersMock{
		GetOrderByIDFunc: func(ctx context.Context, id int) (models.Order, error) {
			return models.Order{}, errors.New("db down")
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/orders/1", "")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =======================
// GetOrdersByUserEmail
// =======================

func TestGetOrdersByUserEmail_Success(t *testing.T) {
	mock := &OrdersMock{
		GetOrdersByUserEmailFunc: func(ctx context.Context, email string) ([]models.Order, error) {
			if email != "ann@example.com" {
				return nil, fmt.Errorf("unexpected email: %s", email)
			}
			return []models.Order{
				{ID: 1, UserID: 1, TotalPrice: "100.00", CreatedAt: time.Now()},
			}, nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/users/ann@example.com/orders", "")

	if w.Code != http.StatusOK {
		t.Fatalf("want code %d, got %d", http.StatusOK, w.Code)
	}

	var got []models.Order
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("unexpected orders: %+v", got)
	}
}

func TestGetOrdersByUserEmail_Fail(t *testing.T) {
	mock := &OrdersMock{
		GetOrdersByUserEmailFunc: func(ctx context.Context, email string) ([]models.Order, error) {
			return nil, errors.New("db down")
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/users/ann@example.com/orders", "")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =======================
// GetOrderItemsByOrderID
// =======================

func TestGetOrderItemsByOrderID_Success(t *testing.T) {
	mock := &OrdersMock{
		GetOrderItemsByOrderIDFunc: func(ctx context.Context, orderID int) ([]models.OrderItem, error) {
			if orderID != 1 {
				return nil, fmt.Errorf("unexpected order id: %d", orderID)
			}
			return []models.OrderItem{
				{ID: 1, OrderID: 1, ProductID: 2, Quantity: 3},
			}, nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/orders/1/items", "")

	if w.Code != http.StatusOK {
		t.Fatalf("want code %d, got %d", http.StatusOK, w.Code)
	}

	var got []models.OrderItem
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(got) != 1 || got[0].ProductID != 2 || got[0].Quantity != 3 {
		t.Fatalf("unexpected order items: %+v", got)
	}
}

// =======================
// GetUserOrderHistory
// =======================

func TestGetUserOrderHistory_Success(t *testing.T) {
	mock := &OrdersMock{
		GetUserOrderHistoryFunc: func(ctx context.Context, email string) ([]models.OrderDetail, error) {
			if email != "ann@example.com" {
				return nil, fmt.Errorf("unexpected email: %s", email)
			}
			return []models.OrderDetail{
				{
					OrderID:           1,
					TransactionStatus: "created",
					Items: []models.OrderItemDetail{
						{ProductID: 2, ProductName: "Dog toy", Quantity: 3, UnitPrice: "9.99"},
					},
				},
			}, nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/users/ann@example.com/history", "")

	if w.Code != http.StatusOK {
		t.Fatalf("want code %d, got %d", http.StatusOK, w.Code)
	}

	var got []models.OrderDetail
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(got) != 1 || got[0].OrderID != 1 || len(got[0].Items) != 1 {
		t.Fatalf("unexpected history: %+v", got)
	}
}

// =======================
// PlaceOrder (checkout)
// =======================

func TestPlaceOrder_Success(t *testing.T) {
	mock := &OrdersMock{
		PlaceOrderFunc: func(ctx context.Context, userEmail string, items []models.OrderItem) (int, error) {
			if userEmail != "ann@example.com" {
				return 0, fmt.Errorf("unexpected email: %s", userEmail)
			}
			if len(items) != 1 || items[0].ProductID != 2 || items[0].Quantity != 3 {
				return 0, fmt.Errorf("unexpected items: %+v", items)
			}
			return 42, nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPost, "/checkout",
		`{"email":"ann@example.com","items":[{"product_id":2,"quantity":3}]}`)

	if w.Code != http.StatusOK {
		t.Fatalf("want code %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if id, ok := resp["order_id"].(float64); !ok || int(id) != 42 {
		t.Fatalf("want order_id 42 in response, got %s", w.Body.String())
	}
}

func TestPlaceOrder_BadRequest(t *testing.T) {
	r := newTestRouter(&OrdersMock{})
	w := performRequest(r, http.MethodPost, "/checkout", `{"email":`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ТЕСТ ПАДАЁТ ПОКА НЕ ИСПРАВЛЕН internal/storage/postgres/order.go:
// при нехватке товара PlaceOrder возвращает обёрнутый ErrNotFound,
// а хендлер для этой ситуации ждёт ErrConflict (409).
func TestPlaceOrder_NotEnoughStock(t *testing.T) {
	mock := &OrdersMock{
		PlaceOrderFunc: func(ctx context.Context, userEmail string, items []models.OrderItem) (int, error) {
			return 0, fmt.Errorf("product 2: %w", storage.ErrNotFound)
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPost, "/checkout",
		`{"email":"ann@example.com","items":[{"product_id":2,"quantity":3}]}`)

	if w.Code != http.StatusConflict {
		t.Fatalf("want code %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestPlaceOrder_UserNotFound(t *testing.T) {
	mock := &OrdersMock{
		PlaceOrderFunc: func(ctx context.Context, userEmail string, items []models.OrderItem) (int, error) {
			return 0, storage.ErrNotFound
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPost, "/checkout",
		`{"email":"ghost@example.com","items":[{"product_id":2,"quantity":3}]}`)

	if w.Code != http.StatusNotFound {
		t.Fatalf("want code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestPlaceOrder_Fail(t *testing.T) {
	mock := &OrdersMock{
		PlaceOrderFunc: func(ctx context.Context, userEmail string, items []models.OrderItem) (int, error) {
			return 0, errors.New("db down")
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPost, "/checkout",
		`{"email":"ann@example.com","items":[{"product_id":2,"quantity":3}]}`)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
