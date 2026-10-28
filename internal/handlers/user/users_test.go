package user

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

func init() {
	gin.SetMode(gin.TestMode)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestRouter(mock *UsersMock) *gin.Engine {
	h := New(testLogger(), mock)

	r := gin.New()
	r.GET("/users", h.GetAllUsers)
	r.GET("/users/:email", h.GetUserByEmail)
	r.POST("/users", h.CreateUser)
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
// GetAllUsers
// =======================

func TestGetAllUsers_Success(t *testing.T) {
	expected := []models.User{
		{ID: 1, Name: "Ann", Email: "ann@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
	}
	mock := &UsersMock{
		GetAllUsersFunc: func(ctx context.Context) ([]models.User, error) {
			return expected, nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/users", "")

	if w.Code != http.StatusOK {
		t.Fatalf("want code %d, got %d", http.StatusOK, w.Code)
	}

	var got []models.User
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(got) != len(expected) {
		t.Fatalf("want %d users, got %d", len(expected), len(got))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("user[%d]: want %+v, got %+v", i, expected[i], got[i])
		}
	}
}

func TestGetAllUsers_Error(t *testing.T) {
	mock := &UsersMock{
		GetAllUsersFunc: func(ctx context.Context) ([]models.User, error) {
			return nil, errors.New("db down")
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/users", "")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =======================
// GetUserByEmail
// =======================

func TestGetUserByEmail_Success(t *testing.T) {
	mock := &UsersMock{
		GetUserByEmailFunc: func(ctx context.Context, email string) (models.User, error) {
			if email != "ann@example.com" {
				return models.User{}, fmt.Errorf("unexpected email: %s", email)
			}
			return models.User{ID: 1, Name: "Ann", Email: "ann@example.com"}, nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/users/ann@example.com", "")

	if w.Code != http.StatusOK {
		t.Fatalf("want code %d, got %d", http.StatusOK, w.Code)
	}

	var got models.User
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.ID != 1 || got.Name != "Ann" || got.Email != "ann@example.com" {
		t.Fatalf("unexpected user: %+v", got)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	mock := &UsersMock{
		GetUserByEmailFunc: func(ctx context.Context, email string) (models.User, error) {
			return models.User{}, storage.ErrNotFound
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/users/ghost@example.com", "")

	if w.Code != http.StatusNotFound {
		t.Fatalf("want code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetUserByEmail_Fail(t *testing.T) {
	mock := &UsersMock{
		GetUserByEmailFunc: func(ctx context.Context, email string) (models.User, error) {
			return models.User{}, errors.New("db down")
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodGet, "/users/ann@example.com", "")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =======================
// CreateUser
// =======================

func TestCreateUser_Success(t *testing.T) {
	mock := &UsersMock{
		CreateUserFunc: func(ctx context.Context, u models.User) error {
			if u.Name != "Ann" || u.Email != "ann@example.com" {
				return fmt.Errorf("unexpected user: %+v", u)
			}
			return nil
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPost, "/users", `{"name":"Ann","email":"ann@example.com"}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("want code %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestCreateUser_BadRequest(t *testing.T) {
	r := newTestRouter(&UsersMock{})
	w := performRequest(r, http.MethodPost, "/users", `{"name":`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ТЕСТ ПАДАЁТ ПОКА НЕ ИСПРАВЛЕН internal/storage/postgres/user.go + handlers/user/users.go:
// при ErrConflict хендлер должен возвращать 409 (сейчас 500).
func TestCreateUser_Conflict(t *testing.T) {
	mock := &UsersMock{
		CreateUserFunc: func(ctx context.Context, u models.User) error {
			return storage.ErrConflict
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPost, "/users", `{"name":"Ann","email":"ann@example.com"}`)

	if w.Code != http.StatusConflict {
		t.Fatalf("want code %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestCreateUser_Fail(t *testing.T) {
	mock := &UsersMock{
		CreateUserFunc: func(ctx context.Context, u models.User) error {
			return errors.New("db down")
		},
	}

	r := newTestRouter(mock)
	w := performRequest(r, http.MethodPost, "/users", `{"name":"Ann","email":"ann@example.com"}`)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
