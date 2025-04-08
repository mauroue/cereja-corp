package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mauroue/cereja-corp/internal/db"
	"github.com/mauroue/cereja-corp/internal/receipts"
)

func main() {
	// Connect to database
	_, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize router
	router := gin.Default()

	// Set HTML templates - load all template files
	router.LoadHTMLGlob("internal/receipts/templates/*.html")
	router.LoadHTMLGlob("internal/receipts/templates/partials/*.html")

	// Add current year to all templates
	router.Use(func(c *gin.Context) {
		c.Set("currentYear", time.Now().Year())
		c.Next()
	})

	// Setup a basic route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Cereja Corp",
		})
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Setup receipt scanner app
	uploadDir := filepath.Join(".", "uploads", "receipts")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Set up API handler
	receiptHandler, err := receipts.NewHandler(uploadDir)
	if err != nil {
		log.Fatalf("Failed to initialize receipt handler: %v", err)
	}
	receiptHandler.RegisterRoutes(router)

	// Set up Web handler
	templatesDir := filepath.Join("internal", "receipts", "templates")
	webHandler, err := receipts.NewWebHandler(receiptHandler, templatesDir)
	if err != nil {
		log.Fatalf("Failed to initialize web handler: %v", err)
	}
	webHandler.RegisterRoutes(router)

	// Start the server
	port := ":8080"
	log.Printf("Starting server on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
