package receipts

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mauroue/cereja-corp/internal/db"
)

// Handler manages HTTP requests for receipts
type Handler struct {
	repo       *Repository
	ocrService *OCRService
}

// NewHandler creates a new receipt handler
func NewHandler(uploadDir string) (*Handler, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	repo := NewRepository(database)

	// Get AWS region from environment variable or use a default
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1" // Default region if not specified
	}

	ocrService := NewOCRService(uploadDir, awsRegion)

	return &Handler{
		repo:       repo,
		ocrService: ocrService,
	}, nil
}

// RegisterRoutes registers the receipt handler routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	receipts := router.Group("/receipts")
	{
		receipts.POST("/upload", h.UploadReceipt)
		receipts.GET("/:id", h.GetReceipt)
		receipts.GET("/:id/items", h.GetReceiptItems)
		receipts.GET("/", h.ListReceipts)
	}
}

// UploadReceipt handles upload of receipt images
func (h *Handler) UploadReceipt(c *gin.Context) {
	// Parse multipart form with 32MB max memory
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	// Get the uploaded file
	file, header, err := c.Request.FormFile("receipt")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Save the image
	imagePath, err := h.ocrService.SaveImage(fileData, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// Process the receipt
	receipt, items, err := h.ocrService.ProcessReceipt(imagePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process receipt"})
		return
	}

	// Ensure the default store exists
	if err := h.repo.EnsureDefaultStore(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ensure store exists"})
		return
	}

	// Save receipt data to database
	receiptID, err := h.repo.CreateReceipt(receipt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save receipt"})
		return
	}

	// Save receipt items
	for _, item := range items {
		item.ReceiptID = receiptID
		_, err := h.repo.CreateReceiptItem(item)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save receipt items"})
			return
		}
	}

	// Return receipt data
	c.JSON(http.StatusOK, gin.H{
		"id":            receiptID,
		"store_name":    receipt.StoreName,
		"purchase_date": receipt.PurchaseDate,
		"total_amount":  receipt.TotalAmount,
		"items_count":   len(items),
	})
}

// GetReceipt handles retrieval of a single receipt
func (h *Handler) GetReceipt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receipt ID"})
		return
	}

	receipt, err := h.repo.GetReceiptByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// GetReceiptItems handles retrieval of items for a receipt
func (h *Handler) GetReceiptItems(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receipt ID"})
		return
	}

	items, err := h.repo.GetReceiptItems(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve receipt items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

// ListReceipts handles listing all receipts with pagination
func (h *Handler) ListReceipts(c *gin.Context) {
	// Implement pagination and filtering here
	c.JSON(http.StatusOK, gin.H{"message": "List receipts - Not implemented yet"})
}
