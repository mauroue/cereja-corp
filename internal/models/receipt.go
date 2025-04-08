package models

import (
	"time"
)

// Receipt represents a purchase receipt with metadata
type Receipt struct {
	ID           int64     `json:"id"`
	StoreID      int64     `json:"store_id"`
	StoreName    string    `json:"store_name"`
	PurchaseDate time.Time `json:"purchase_date"`
	TotalAmount  float64   `json:"total_amount"`
	ImagePath    string    `json:"image_path"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ReceiptItem represents an individual item from a purchase receipt
type ReceiptItem struct {
	ID          int64     `json:"id"`
	ReceiptID   int64     `json:"receipt_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Quantity    float64   `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	TotalPrice  float64   `json:"total_price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Store represents a store where purchases are made
type Store struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
