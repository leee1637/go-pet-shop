package models

import "time"

type Product struct {
	ID    int
	Name  string
	Price float64
	Stock int // количество на складе
}

type User struct {
	ID    int
	Name  string
	Email string
}

type Order struct {
	ID         int
	UserID     int
	TotalPrice string
	CreatedAt  time.Time
}

type OrderItem struct {
	ID        int
	OrderID   int
	ProductID int
	Quantity  int
}

type PopularProduct struct {
	ProductID   int
	ProductName string
	TotalSold   int
}

type OrderDetail struct {
	OrderID           int
	CreatedAt         time.Time
	TotalPrice        string
	TransactionStatus string
	Items             []OrderItemDetail
}

type OrderItemDetail struct {
	ProductID   int
	ProductName string
	Quantity    int
	UnitPrice   string
}
