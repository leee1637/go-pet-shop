package user

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Users interface {
	CreateUser(ctx context.Context, user models.User) error
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type Handler struct {
	log     *slog.Logger
	storage Users
}

func New(log *slog.Logger, storage Users) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
	}

}

type createRequsetUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *Handler) CreateUser(g *gin.Context) {
	var createUser createRequsetUser
	err := g.ShouldBindJSON(&createUser)
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Name:  createUser.Name,
		Email: createUser.Email,
	}

	err = h.storage.CreateUser(g.Request.Context(), user)
	if err != nil {
		if errors.Is(err, storage.ErrConflict) {
			g.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		h.log.Error("failed to create user", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	g.Status(http.StatusCreated)
}

func (h *Handler) GetUserByEmail(g *gin.Context) {
	email := g.Param("email")
	if email == "" {
		g.JSON(http.StatusBadRequest, gin.H{"error": "cant email bee nil"})
		return
	}

	user, err := h.storage.GetUserByEmail(g.Request.Context(), email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			g.JSON(http.StatusNotFound, gin.H{"error": "not found user"})
			return
		}

		h.log.Error("failed to get user", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	g.JSON(http.StatusOK, user)
}

func (h *Handler) GetAllUsers(g *gin.Context) {
	user, err := h.storage.GetAllUsers(g.Request.Context())
	if err != nil {
		h.log.Error("fail get product", slog.String("error", err.Error()))
		g.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	g.JSON(http.StatusOK, user)
}
