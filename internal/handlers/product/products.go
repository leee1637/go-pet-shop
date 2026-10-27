package product

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

type Products interface {
	GetAllProducts(ctx context.Context) ([]models.Product, error)
	CreateProduct(ctx context.Context, product models.Product) (int, error)
	DeleteProduct(ctx context.Context, id int) error
	UpdateProduct(ctx context.Context, product models.Product) error
	GetProductByID(ctx context.Context, id int) (models.Product, error)

	GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error)
}

type Handler struct {
	log     *slog.Logger
	storage Products
}

func New(log *slog.Logger, storage Products) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
	}
}

func (h *Handler) GetAllProducts(g *gin.Context) {
	products, err := h.storage.GetAllProducts(g.Request.Context())
	if err != nil {
		h.log.Error("failed get all products", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	g.JSON(http.StatusOK, products)
}

func (h *Handler) GetPopularProducts(g *gin.Context) {
	products, err := h.storage.GetPopularProducts(g.Request.Context())
	if err != nil {
		h.log.Error("failed get popular products", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	g.JSON(http.StatusOK, products)
}

func (h *Handler) GetProductByID(g *gin.Context) {
	id, err := strconv.Atoi(g.Param("id"))
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.storage.GetProductByID(g.Request.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			g.JSON(http.StatusNotFound, gin.H{"error": "not found product"})
			return
		}

		h.log.Error("failed to get product", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	g.JSON(http.StatusOK, p)
}

type createProductRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func (h *Handler) CreateProduct(g *gin.Context) {
	var p createProductRequest
	err := g.ShouldBindJSON(&p)
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{
		Name:  p.Name,
		Price: p.Price,
		Stock: p.Stock,
	}

	productID, err := h.storage.CreateProduct(g.Request.Context(), product)
	if err != nil {
		h.log.Error("failed to create product", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	g.JSON(http.StatusCreated, gin.H{"id": productID})

}

func (h *Handler) DeleteProduct(g *gin.Context) {
	id, err := strconv.Atoi(g.Param("id"))
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.storage.DeleteProduct(g.Request.Context(), id)
	if err != nil {
		h.log.Error("failed to delete product", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}

	g.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *Handler) UpdateProduct(g *gin.Context) {
	var p createProductRequest
	err := g.ShouldBindJSON(&p)
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := strconv.Atoi(g.Param("id"))
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{
		ID:    id,
		Name:  p.Name,
		Price: p.Price,
		Stock: p.Stock,
	}

	err = h.storage.UpdateProduct(g.Request.Context(), product)
	if err != nil {
		h.log.Error("failed to create product", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	g.JSON(http.StatusOK, gin.H{"id": id})
}
