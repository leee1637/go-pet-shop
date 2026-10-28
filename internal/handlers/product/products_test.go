package product

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

	"github.com/gin-gonic/gin"
)

// init переключает gin в тестовый режим, чтобы не мешать логи.
func init() {
	gin.SetMode(gin.TestMode)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newTestRouter собирает тот же роут, что и main.go, но с моком вместо реального storage.
func newTestRouter(mock *ProductsMock) *gin.Engine {
	h := New(testLogger(), mock)

	r := gin.New()
	r.GET("/products", h.GetAllProducts)
	r.GET("/products/popular", h.GetPopularProducts)
	r.GET("/products/:id", h.GetProductByID)
	r.POST("/products", h.CreateProduct)
	r.PUT("/products/:id", h.UpdateProduct)
	r.DELETE("/products/:id", h.DeleteProduct)
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
// GetAllProducts
// =======================

func TestGetAllProducts_Success(t *testing.T) {
	expected := []models.Product{
		{ID: 1, Name: "Dog toy", Price: 9.99, Stock: 15},
		{ID: 2, Name: "Cat food", Price: 19.5, Stock: 3},
	}
	mock := &ProductsMock{
		GetAllProductsFunc: func(ctx context.Context) ([]models.Product, error) {
			return expected, nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/products", "")

	if w.Code != http.StatusOK {
		t.Fatalf("want code %d, got %d", http.StatusOK, w.Code)
	}

	var got []models.Product
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(got) != len(expected) {
		t.Fatalf("want %d products, got %d", len(expected), len(got))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("product[%d]: want %+v, got %+v", i, expected[i], got[i])
		}
	}
}

func TestGetAllProducts_Error(t *testing.T) {
	mock := &ProductsMock{
		GetAllProductsFunc: func(ctx context.Context) ([]models.Product, error) {
			return nil, errors.New("db down")
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/products", "")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =======================
// GetProductByID
// =======================

func TestGetProductByID_Success(t *testing.T) {
	mock := &ProductsMock{
		GetProductByIDFunc: func(ctx context.Context, id int) (models.Product, error) {
			if id != 1 {
				return models.Product{}, fmt.Errorf("unexpected id: %d", id)
			}
			return models.Product{ID: 1, Name: "Dog toy", Price: 9.99, Stock: 15}, nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/products/1", "")

	if w.Code != http.StatusOK {
		t.Fatalf("want code %d, got %d", http.StatusOK, w.Code)
	}

	var got models.Product
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.ID != 1 || got.Name != "Dog toy" {
		t.Fatalf("want Dog toy product, got %+v", got)
	}
}

func TestGetProductByID_NotFound(t *testing.T) {
	mock := &ProductsMock{
		GetProductByIDFunc: func(ctx context.Context, id int) (models.Product, error) {
			return models.Product{}, storage.ErrNotFound
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/products/999", "")

	if w.Code != http.StatusNotFound {
		t.Fatalf("want code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetProductByID_BadID(t *testing.T) {
	r := newTestRouter(&ProductsMock{})
	w := performRequest(r, http.MethodGet, "/products/not-a-number", "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// =======================
// CreateProduct
// =======================

func TestCreateProduct_Success(t *testing.T) {
	mock := &ProductsMock{
		CreateProductFunc: func(ctx context.Context, p models.Product) (int, error) {
			if p.Name != "Dog toy" || p.Price != 9.99 || p.Stock != 15 {
				return 0, fmt.Errorf("unexpected product: %+v", p)
			}
			return 42, nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPost, "/products", `{"name":"Dog toy","price":9.99,"stock":15}`)

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

func TestCreateProduct_BadRequest(t *testing.T) {
	r := newTestRouter(&ProductsMock{})
	w := performRequest(r, http.MethodPost, "/products", `{"name":`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateProduct_Fail(t *testing.T) {
	mock := &ProductsMock{
		CreateProductFunc: func(ctx context.Context, p models.Product) (int, error) {
			return 0, errors.New("db down")
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPost, "/products", `{"name":"Dog toy","price":9.99,"stock":15}`)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =======================
// UpdateProduct
// =======================

func TestUpdateProduct_Success(t *testing.T) {
	mock := &ProductsMock{
		UpdateProductFunc: func(ctx context.Context, p models.Product) error {
			if p.ID != 1 || p.Name != "Dog toy updated" || p.Price != 12.5 || p.Stock != 10 {
				return fmt.Errorf("unexpected product: %+v", p)
			}
			return nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPut, "/products/1", `{"name":"Dog toy updated","price":12.5,"stock":10}`)

	if w.Code != http.StatusOK {
		t.Fatalf("want code %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestUpdateProduct_BadRequest(t *testing.T) {
	r := newTestRouter(&ProductsMock{})
	w := performRequest(r, http.MethodPut, "/products/1", `{"name":`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateProduct_BadID(t *testing.T) {
	r := newTestRouter(&ProductsMock{})
	w := performRequest(r, http.MethodPut, "/products/not-a-number", `{"name":"x","price":1,"stock":1}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateProduct_Fail(t *testing.T) {
	mock := &ProductsMock{
		UpdateProductFunc: func(ctx context.Context, p models.Product) error {
			return errors.New("db down")
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPut, "/products/1", `{"name":"Dog toy updated","price":12.5,"stock":10}`)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =======================
// DeleteProduct
// =======================

func TestDeleteProduct_Success(t *testing.T) {
	mock := &ProductsMock{
		DeleteProductFunc: func(ctx context.Context, id int) error {
			if id != 1 {
				return fmt.Errorf("unexpected id: %d", id)
			}
			return nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodDelete, "/products/1", "")

	if w.Code != http.StatusOK {
		t.Fatalf("want code %d, got %d", http.StatusOK, w.Code)
	}
}

// ТЕСТ ПАДАЁТ ПОКА НЕ ИСПРАВЛЕН handlers/product/products.go:
// в DeleteProduct при ошибке нет `return` после 500-ответа — клиент получает "deleted".
func TestDeleteProduct_Fail(t *testing.T) {
	mock := &ProductsMock{
		DeleteProductFunc: func(ctx context.Context, id int) error {
			return storage.ErrNotFound
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodDelete, "/products/999", "")

	if w.Code != http.StatusNotFound {
		t.Fatalf("want code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteProduct_BadID(t *testing.T) {
	r := newTestRouter(&ProductsMock{})
	w := performRequest(r, http.MethodDelete, "/products/not-a-number", "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// =======================
// GetPopularProducts
// =======================

func TestGetPopularProducts_Success(t *testing.T) {
	expected := []models.PopularProduct{
		{ProductID: 1, ProductName: "Dog toy", TotalSold: 100},
	}
	mock := &ProductsMock{
		GetPopularProductsFunc: func(ctx context.Context) ([]models.PopularProduct, error) {
			return expected, nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/products/popular", "")

	if w.Code != http.StatusOK {
		t.Fatalf("want code %d, got %d", http.StatusOK, w.Code)
	}

	var got []models.PopularProduct
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(got) != 1 || got[0].TotalSold != 100 {
		t.Fatalf("unexpected response: %+v", got)
	}
}

func TestGetPopularProducts_Error(t *testing.T) {
	mock := &ProductsMock{
		GetPopularProductsFunc: func(ctx context.Context) ([]models.PopularProduct, error) {
			return nil, errors.New("db down")
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/products/popular", "")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
