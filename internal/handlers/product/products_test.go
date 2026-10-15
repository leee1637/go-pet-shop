package product

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-pet-shop/internal/handlers/product/mocks"
	"go-pet-shop/internal/models"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	testifymock "github.com/stretchr/testify/mock"
)

// Get Product - Ready
func TestGetAllProducts_Success(t *testing.T) {
	// Мокаем storage — он вернёт один продукт.
	mock := mocks.NewProducts(t)
	mock.On("GetAllProducts", testifymock.Anything).Return([]models.Product{
		{ID: 1, Name: "Dog Food"},
	}, nil)

	// Создаем HTTP-запрос GET /products
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	// Создаем хендлер с мок-хранилищем
	handler := New(slog.Default(), mock)

	// Вызываем метод GetAllProducts, который является http.HandlerFunc
	handler.GetAllProducts(w, req)

	// Проверяем HTTP-код
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
func TestGetAllProducts_Error(t *testing.T) {
	// Мокаем storage — он будет возвращать ошибку
	mock := mocks.NewProducts(t)
	mock.On("GetAllProducts", testifymock.Anything).Return(nil, errors.New("DB error"))

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)
	handler.GetAllProducts(w, req)

	// Ожидаем HTTP 500
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetAllProducts(t *testing.T) { // для себя дублировал код
	tests := []struct {
		name       string
		p          *mocks.Products
		wantStatus int
	}{{
		name: "Success",
		p: func() *mocks.Products {
			m := mocks.NewProducts(t)
			m.On("GetAllProducts", testifymock.Anything).Return([]models.Product{{ID: 1, Name: "Dog Food"}}, nil)
			return m
		}(),
		wantStatus: 200,
	},
		{
			name: "Error",
			p: func() *mocks.Products {
				m := mocks.NewProducts(t)
				m.On("GetAllProducts", testifymock.Anything).Return(nil, errors.New("DB error"))
				return m
			}(),
			wantStatus: 500,
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			w := httptest.NewRecorder()

			handler := New(slog.Default(), tt.p)
			handler.GetAllProducts(w, req)

			if tt.wantStatus != w.Code {
				t.Errorf("expected %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// =======================
// Create Product
// =======================

func TestCreateProduct_Success(t *testing.T) {
	// TODO: Написать Unit-тест для создания продукта (200 OK)
	mock := mocks.NewProducts(t)
	mock.On("CreateProduct", testifymock.Anything, testifymock.Anything).Return(1, nil)
	product := models.Product{ID: 1, Name: "Помидор", Price: 32.32, Stock: 214214}

	b, _ := json.Marshal(product)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(b))

	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)

	handler.CreateProduct(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["status"] != "Product created successfully" {
		t.Errorf("unexpected status field: %v", resp["status"])
	}

	if resp["id"] != 1.0 { // id в JSON — это float64, потому что map[string]any
		t.Errorf("expected id 1, got %v", resp["id"])
	}
}

func TestCreateProduct_BadRequest(t *testing.T) {
	// TODO: Написать Unit-тест для создания продукта с невалидным JSON (400 Bad Request)
	mock := mocks.NewProducts(t)

	product := models.Product{ID: 23, Name: "", Stock: -2333}

	b, _ := json.Marshal(product)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(b))

	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)

	handler.CreateProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	var resp map[string]any

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["error"] != "Bad request" {
		t.Errorf("unexpected status field: %v", resp["error"])
	}

}

func TestCreateProduct_Fail(t *testing.T) {
	// TODO: Написать Unit-тест для создания продукта при ошибке сервиса (500 Internal Server Error)
	mock := mocks.NewProducts(t)
	mock.On("CreateProduct", testifymock.Anything, testifymock.Anything).Return(0, errors.New("fail bd"))
	product := models.Product{ID: 1, Name: "Помидор", Price: 32.32, Stock: 214214}

	b, _ := json.Marshal(product)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(b))

	w := httptest.NewRecorder()

	handler := New(slog.Default(), mock)

	handler.CreateProduct(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var resp map[string]any

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["error"] != "Internal server error" {
		t.Errorf("unexpected status field: %v", resp["error"])
	}

	if resp["message"] != "Failed to create product" {
		t.Errorf("unexpected status field: %v", resp["message"])
	}
}

// =======================
// Update Product
// =======================

func TestUpdateProduct(t *testing.T) {

	test := []struct {
		name       string
		mock       *mocks.Products
		wantStatus int
	}{
		{
			name:       "Success",
			wantStatus: 200,
			mock: func() *mocks.Products {
				m := mocks.NewProducts(t)
				m.On("UpdateProduct", testifymock.Anything, testifymock.Anything).Return(nil)
				return m
			}()},
		{
			name:       "BadRequest",
			wantStatus: 400,
			mock:       mocks.NewProducts(t)},
		{
			name:       "Fail",
			wantStatus: 500,
			mock: func() *mocks.Products {
				m := mocks.NewProducts(t)
				m.On("UpdateProduct", testifymock.Anything, testifymock.Anything).Return(errors.New("fail bd"))
				return m
			}()}}

	for _, tt := range test {

		t.Run(tt.name, func(t *testing.T) {
			product := models.Product{Name: "Помидор", Price: 32.32, Stock: 214214}
			b, _ := json.Marshal(product)
			req := httptest.NewRequest(http.MethodPut, "/products/1", bytes.NewReader(b))
			rctx := chi.NewRouteContext()
			if tt.wantStatus != 400 {
				rctx.URLParams.Add("id", "1")
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()
			handler := New(slog.Default(), tt.mock)
			handler.UpdateProduct(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, w.Code)
			}

			var resp map[string]any

			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			switch tt.wantStatus {
			case 200:
				if resp["status"] != "Product updated successfully" {
					t.Errorf("unexpected status field: %v", resp["status"])
				}
			case 400:
				if resp["error"] != "Bad request" {
					t.Errorf("unexpected error field: %v", resp["error"])
				}
			case 500:
				if resp["error"] != "Internal server error" {
					t.Errorf("unexpected error field: %v", resp["error"])
				}
			}

		})
	}

}

// =======================
// Delete Product
// =======================

func TestDeleteProduct(t *testing.T) {
	test := []struct {
		name       string
		mock       *mocks.Products
		wantStatus int
	}{
		{
			name:       "Success",
			wantStatus: 200,
			mock: func() *mocks.Products {
				m := mocks.NewProducts(t)
				m.On("DeleteProduct", testifymock.Anything, 1).Return(nil)
				return m
			}()},
		{
			name:       "BadRequest",
			wantStatus: 400,
			mock:       mocks.NewProducts(t)},
		{
			name:       "Fail",
			wantStatus: 500,
			mock: func() *mocks.Products {
				m := mocks.NewProducts(t)
				m.On("DeleteProduct", testifymock.Anything, 1).Return(errors.New("fail bd"))
				return m
			}()}}

	for _, tt := range test {

		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/products/1", nil)
			rctx := chi.NewRouteContext()
			if tt.wantStatus != 400 {
				rctx.URLParams.Add("id", "1")
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()
			handler := New(slog.Default(), tt.mock)
			handler.DeleteProduct(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, w.Code)
			}

			var resp map[string]any

			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			switch tt.wantStatus {
			case 200:
				if resp["status"] != "Product deleted successfully" {
					t.Errorf("unexpected status field: %v", resp["status"])
				}
			case 400:
				if resp["error"] != "Bad request" {
					t.Errorf("unexpected error field: %v", resp["error"])
				}
			case 500:
				if resp["error"] != "Internal server error" {
					t.Errorf("unexpected error field: %v", resp["error"])
				}
			}

		})
	}
}
