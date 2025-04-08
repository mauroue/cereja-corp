package receipts

import (
	"database/sql"
	"time"

	"github.com/mauroue/cereja-corp/internal/models"
)

// Repository handles database operations for receipts
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new receipt repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateReceipt inserts a new receipt into the database
func (r *Repository) CreateReceipt(receipt *models.Receipt) (int64, error) {
	query := `
		INSERT INTO receipts (store_id, store_name, purchase_date, total_amount, image_path, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	now := time.Now()
	receipt.CreatedAt = now
	receipt.UpdatedAt = now

	var id int64
	err := r.db.QueryRow(
		query,
		receipt.StoreID,
		receipt.StoreName,
		receipt.PurchaseDate,
		receipt.TotalAmount,
		receipt.ImagePath,
		receipt.CreatedAt,
		receipt.UpdatedAt,
	).Scan(&id)

	return id, err
}

// CreateReceiptItem inserts a new receipt item into the database
func (r *Repository) CreateReceiptItem(item *models.ReceiptItem) (int64, error) {
	query := `
		INSERT INTO receipt_items (receipt_id, name, description, quantity, unit_price, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	var id int64
	err := r.db.QueryRow(
		query,
		item.ReceiptID,
		item.Name,
		item.Description,
		item.Quantity,
		item.UnitPrice,
		item.TotalPrice,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&id)

	return id, err
}

// GetReceiptByID retrieves a receipt by its ID
func (r *Repository) GetReceiptByID(id int64) (*models.Receipt, error) {
	query := `
		SELECT id, store_id, store_name, purchase_date, total_amount, image_path, created_at, updated_at
		FROM receipts
		WHERE id = $1
	`

	var receipt models.Receipt
	err := r.db.QueryRow(query, id).Scan(
		&receipt.ID,
		&receipt.StoreID,
		&receipt.StoreName,
		&receipt.PurchaseDate,
		&receipt.TotalAmount,
		&receipt.ImagePath,
		&receipt.CreatedAt,
		&receipt.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &receipt, nil
}

// GetReceiptItems retrieves all items for a specific receipt
func (r *Repository) GetReceiptItems(receiptID int64) ([]*models.ReceiptItem, error) {
	query := `
		SELECT id, receipt_id, name, description, quantity, unit_price, total_price, created_at, updated_at
		FROM receipt_items
		WHERE receipt_id = $1
		ORDER BY id
	`

	rows, err := r.db.Query(query, receiptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.ReceiptItem
	for rows.Next() {
		var item models.ReceiptItem
		if err := rows.Scan(
			&item.ID,
			&item.ReceiptID,
			&item.Name,
			&item.Description,
			&item.Quantity,
			&item.UnitPrice,
			&item.TotalPrice,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, rows.Err()
}

// EnsureDefaultStore creates a default store with ID 1 if it doesn't exist
func (r *Repository) EnsureDefaultStore() error {
	query := `
		INSERT INTO stores (id, name, address, created_at, updated_at)
		VALUES (1, 'Sample Store', '123 Sample St, Sample City', $1, $2)
		ON CONFLICT (id) DO NOTHING
	`

	now := time.Now()
	_, err := r.db.Exec(query, now, now)
	return err
}

// ListReceipts retrieves receipts with pagination
func (r *Repository) ListReceipts(page, pageSize int, search string) ([]*models.Receipt, error) {
	// Default pagination values if not provided
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var query string
	var args []interface{}

	if search != "" {
		query = `
			SELECT id, store_id, store_name, purchase_date, total_amount, image_path, created_at, updated_at
			FROM receipts
			WHERE store_name ILIKE $1
			ORDER BY purchase_date DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{"%" + search + "%", pageSize, offset}
	} else {
		query = `
			SELECT id, store_id, store_name, purchase_date, total_amount, image_path, created_at, updated_at
			FROM receipts
			ORDER BY purchase_date DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{pageSize, offset}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var receipts []*models.Receipt
	for rows.Next() {
		var receipt models.Receipt
		if err := rows.Scan(
			&receipt.ID,
			&receipt.StoreID,
			&receipt.StoreName,
			&receipt.PurchaseDate,
			&receipt.TotalAmount,
			&receipt.ImagePath,
			&receipt.CreatedAt,
			&receipt.UpdatedAt,
		); err != nil {
			return nil, err
		}
		receipts = append(receipts, &receipt)
	}

	return receipts, rows.Err()
}

// GetReceiptsCount returns the total number of receipts matching the search criteria
func (r *Repository) GetReceiptsCount(search string) (int, error) {
	var query string
	var args []interface{}

	if search != "" {
		query = `SELECT COUNT(*) FROM receipts WHERE store_name ILIKE $1`
		args = []interface{}{"%" + search + "%"}
	} else {
		query = `SELECT COUNT(*) FROM receipts`
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)

	return count, err
}
