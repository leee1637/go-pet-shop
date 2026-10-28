package main

import (
	"context"
	"go-pet-shop/internal/config"
	"go-pet-shop/internal/handlers"
	"go-pet-shop/internal/handlers/order"
	"go-pet-shop/internal/handlers/product"
	"go-pet-shop/internal/handlers/user"
	"go-pet-shop/internal/lib/logger"
	"go-pet-shop/internal/storage/postgres"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.MustLoad()

	// Settings logger
	log := logger.SetupLogger(cfg.Env)
	log.Info("starting the project...", slog.String("env", cfg.Env))

	// Settings and started database
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DatabaseTimeout)
	defer cancel()

	storage, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to init storage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			log.Error("failed to close storage", slog.String("error", err.Error()))
		}
		log.Info("storage closed")
	}()

	// Init router
	router := gin.New()
	// Middlewares
	router.Use(gin.Recovery())
	router.Use(handlers.GinLogger(log))

	// Handlers
	// Метод	URL	Описание
	// POST	/users	Создать пользователя
	// GET	/users	Получить список пользователей
	// GET	/users/{email}	Получить пользователя по email
	// GET	/users/{email}/orders	Получить заказы пользователя
	// POST	/products	Добавить новый товар
	// GET	/products	Получить список товаров
	// GET	/products/{id}	Получить товар по ID
	// PUT	/products/{id}	Обновить товар
	// DELETE	/products/{id}	Удалить товар
	// POST	/orders	Создать заказ
	// POST	/orders/{id}/items	Добавить позицию к заказу
	// GET	/orders/{id}	Получить детали заказа
	// POST	/checkout	Оформление заказа (транзакция)
	// GET	/users/{email}/history	История заказов и транзакций
	// GET	/products/popular	Популярные товары (аналитика)
	// GET	/health	Проверка состояния сервера

	productsHandler := product.New(log, storage)
	usersHandler := user.New(log, storage)
	ordersHandler := order.New(log, storage)

	router.GET("/health", handlers.StatusHandler)

	router.GET("/users", usersHandler.GetAllUsers)
	router.GET("/users/:email", usersHandler.GetUserByEmail)
	router.GET("/users/:email/orders", ordersHandler.GetOrdersByUserEmail)
	router.POST("/users", usersHandler.CreateUser)
	router.POST("/products", productsHandler.CreateProduct)
	router.GET("/products", productsHandler.GetAllProducts)
	router.GET("/products/:id", productsHandler.GetProductByID)
	router.PUT("/products/:id", productsHandler.UpdateProduct)
	router.DELETE("/products/:id", productsHandler.DeleteProduct)
	router.POST("/orders", ordersHandler.CreateOrder)
	router.POST("/orders/:id/items", ordersHandler.AddOrderItem)
	router.GET("/orders/:id/items", ordersHandler.GetOrderItemsByOrderID)
	router.GET("/orders/:id", ordersHandler.GetOrderByID)
	router.POST("/checkout", ordersHandler.PlaceOrder)
	router.GET("/users/:email/history", ordersHandler.GetUserOrderHistory)
	router.GET("/products/popular", productsHandler.GetPopularProducts)

	// Settings and started server
	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	// Запуск сервера в горутине для graceful shutdown
	go func() {
		log.Info("Starting server on", slog.String("address", cfg.HTTPServer.Address))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop // Ждем сигнал завершения

	log.Info("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", slog.String("error", err.Error()))
	}

	log.Info("Server stopped gracefully")
}
