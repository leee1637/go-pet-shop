package order

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Orders interface {
	CreateOrder(ctx context.Context, order models.Order) (int, error)
	AddOrderItem(ctx context.Context, orderItem models.OrderItem) error
	GetOrderByID(ctx context.Context, id int) (models.Order, error)
	GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error)

	GetUserOrderHistory(ctx context.Context, email string) ([]models.OrderDetail, error)

	PlaceOrder(ctx context.Context, userEmail string, items []models.OrderItem) (int, error)
}

type Handler struct {
	log     *slog.Logger
	storage Orders
}

func New(log *slog.Logger, storage Orders) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
	}

}

type createRequsetOrder struct {
	UserID     int    `json:"user_id"`
	TotalPrice string `json:"total_price"`
}

func (h *Handler) CreateOrder(g *gin.Context) {
	var createOrder createRequsetOrder
	err := g.ShouldBindJSON(&createOrder)
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order := models.Order{
		UserID:     createOrder.UserID,
		TotalPrice: createOrder.TotalPrice,
	}

	id, err := h.storage.CreateOrder(g.Request.Context(), order)
	if err != nil {
		h.log.Error("failed to create order", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	g.JSON(http.StatusCreated, gin.H{"id": id})
}

type createRequsetOrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func (h *Handler) AddOrderItem(g *gin.Context) {
	id, err := strconv.Atoi(g.Param("id"))
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var createOrder createRequsetOrderItem
	err = g.ShouldBindJSON(&createOrder)
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderItem := models.OrderItem{
		OrderID:   id,
		ProductID: createOrder.ProductID,
		Quantity:  createOrder.Quantity,
	}

	err = h.storage.AddOrderItem(g.Request.Context(), orderItem)
	if err != nil {
		h.log.Error("failed to add order item", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	g.Status(http.StatusOK)
}

func (h *Handler) GetOrderByID(g *gin.Context) {
	id, err := strconv.Atoi(g.Param("id"))
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.storage.GetOrderByID(g.Request.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			g.JSON(http.StatusNotFound, gin.H{"error": "not found order"})
			return
		}

		h.log.Error("failed to get order", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	g.JSON(http.StatusOK, order)
}

func (h *Handler) GetOrdersByUserEmail(g *gin.Context) {
	email := g.Param("email")
	if email == "" {
		g.JSON(http.StatusBadRequest, gin.H{"error": "cant email bee nil"})
		return
	}

	order, err := h.storage.GetOrdersByUserEmail(g.Request.Context(), email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			g.JSON(http.StatusNotFound, gin.H{"error": "not found order"})
			return
		}

		h.log.Error("failed to get order", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	g.JSON(http.StatusOK, order)
}

func (h *Handler) GetOrderItemsByOrderID(g *gin.Context) {
	id, err := strconv.Atoi(g.Param("id"))
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderItems, err := h.storage.GetOrderItemsByOrderID(g.Request.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			g.JSON(http.StatusNotFound, gin.H{"error": "not found order items"})
			return
		}

		h.log.Error("failed to get order items", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	g.JSON(http.StatusOK, orderItems)
}

func (h *Handler) GetUserOrderHistory(g *gin.Context) {
	email := g.Param("email")
	if email == "" {
		g.JSON(http.StatusBadRequest, gin.H{"error": "cant email bee nil"})
		return
	}

	user, err := h.storage.GetUserOrderHistory(g.Request.Context(), email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			g.JSON(http.StatusNotFound, gin.H{"error": "not found user"})
			return
		}

		h.log.Error("failed to get user order history", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	g.JSON(http.StatusOK, user)
}

type requestPlaceOrder struct {
	Email string `json:"email"`
	Items []struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	} `json:"items"`
}

func (h *Handler) PlaceOrder(g *gin.Context) {
	var req requestPlaceOrder

	err := g.ShouldBindJSON(&req)
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]models.OrderItem, len(req.Items))
	for k, v := range req.Items {
		items[k] = models.OrderItem{
			ProductID: v.ProductID,
			Quantity:  v.Quantity,
		}
	}

	orderID, err := h.storage.PlaceOrder(g.Request.Context(), req.Email, items)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			g.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, storage.ErrConflict):
			g.JSON(http.StatusConflict, gin.H{"error": "insufficient stock"})
		case errors.Is(err, storage.ErrInvalidInput):
			g.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		default:
			h.log.Error("failed to place order", slog.String("error", err.Error()))
			g.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	g.JSON(http.StatusOK, gin.H{"order_id": orderID})
}
