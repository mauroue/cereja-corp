package receipts

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/mauroue/cereja-corp/internal/models"
)

// Global AWS session cache
var (
	awsSessionCache     *session.Session
	awsSessionCacheLock sync.RWMutex
)

// getOrCreateAWSSession creates or retrieves a cached AWS session
func getOrCreateAWSSession(region string) (*session.Session, error) {
	// Check if we have a cached session first
	awsSessionCacheLock.RLock()
	if awsSessionCache != nil {
		sess := awsSessionCache
		awsSessionCacheLock.RUnlock()
		return sess, nil
	}
	awsSessionCacheLock.RUnlock()

	// No cached session, create a new one with a write lock
	awsSessionCacheLock.Lock()
	defer awsSessionCacheLock.Unlock()

	// Double-check in case another goroutine created the session while we were waiting
	if awsSessionCache != nil {
		return awsSessionCache, nil
	}

	// Set a default region if empty
	if region == "" {
		region = "us-east-1"
		fmt.Println("AWS region not specified, using default: us-east-1")
	} else {
		fmt.Printf("Using AWS region: %s\n", region)
	}

	// Create the session with better error handling
	sessionOptions := session.Options{
		Config: aws.Config{
			Region: aws.String(region),
		},
		SharedConfigState: session.SharedConfigEnable,
	}

	sess, err := session.NewSessionWithOptions(sessionOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Validate credentials
	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return nil, fmt.Errorf("invalid AWS credentials: %w", err)
	}

	// Store in cache for future use
	awsSessionCache = sess
	fmt.Println("AWS session created and cached successfully")

	return sess, nil
}

// OCRService handles the optical character recognition for receipts
type OCRService struct {
	uploadDir      string
	awsRegion      string
	textractClient *textract.Textract
}

// NewOCRService creates a new OCR service
func NewOCRService(uploadDir string, awsRegion string) *OCRService {
	var textractClient *textract.Textract

	// Check if AWS credentials are set
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if accessKey == "" || secretKey == "" {
		fmt.Println("AWS credentials not properly configured. AWS_ACCESS_KEY_ID or AWS_SECRET_ACCESS_KEY environment variables are missing.")
	} else {
		// Try to get or create AWS session
		sess, err := getOrCreateAWSSession(awsRegion)
		if err != nil {
			fmt.Printf("AWS session error: %v\n", err)
		} else {
			textractClient = textract.New(sess)
		}
	}

	return &OCRService{
		uploadDir:      uploadDir,
		awsRegion:      awsRegion,
		textractClient: textractClient,
	}
}

// ProcessReceipt processes a receipt image and extracts information
func (s *OCRService) ProcessReceipt(imagePath string) (*models.Receipt, []*models.ReceiptItem, error) {
	// If Textract client is available, use AWS Textract
	if s.textractClient != nil {
		return s.processWithTextract(imagePath)
	}

	// If Textract client is not available, return an error
	return nil, nil, fmt.Errorf("AWS Textract client not available: please configure AWS credentials")
}

// processWithTextract processes the receipt using AWS Textract
func (s *OCRService) processWithTextract(imagePath string) (*models.Receipt, []*models.ReceiptItem, error) {
	// Read the image file
	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read image file: %w", err)
	}

	// Call AWS Textract to analyze the receipt
	input := &textract.AnalyzeExpenseInput{
		Document: &textract.Document{
			Bytes: imageBytes,
		},
	}

	result, err := s.textractClient.AnalyzeExpense(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to analyze receipt with AWS Textract: %w", err)
	}

	// Parse the Textract result into our data structures
	receipt, items, err := s.parseTextractResult(result, imagePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse Textract result: %w", err)
	}

	return receipt, items, nil
}

// parseTextractResult extracts structured data from Textract AnalyzeExpense result
func (s *OCRService) parseTextractResult(result *textract.AnalyzeExpenseOutput, imagePath string) (*models.Receipt, []*models.ReceiptItem, error) {
	receipt := &models.Receipt{
		StoreID:      1, // Default store ID - should be determined by matching vendor name
		StoreName:    "Unknown Store",
		PurchaseDate: time.Now(),
		TotalAmount:  0.0,
		ImagePath:    imagePath,
	}

	var items []*models.ReceiptItem

	// Process each expense document
	for _, doc := range result.ExpenseDocuments {
		// Extract invoice/receipt details
		for _, field := range doc.SummaryFields {
			fieldType := aws.StringValue(field.Type.Text)
			fieldValue := aws.StringValue(field.ValueDetection.Text)

			switch fieldType {
			case "VENDOR_NAME":
				receipt.StoreName = fieldValue
			case "INVOICE_RECEIPT_DATE":
				if date, err := parseDate(fieldValue); err == nil {
					receipt.PurchaseDate = date
				}
			case "TOTAL":
				if total, err := parseFloat(fieldValue); err == nil {
					receipt.TotalAmount = total
				}
			}
		}

		// Extract line items
		for _, table := range doc.LineItemGroups {
			for _, lineItem := range table.LineItems {
				item := &models.ReceiptItem{
					Name:        "Unknown Item",
					Description: "",
					Quantity:    1.0,
					UnitPrice:   0.0,
					TotalPrice:  0.0,
				}

				// Process each field in the line item
				for _, field := range lineItem.LineItemExpenseFields {
					fieldType := aws.StringValue(field.Type.Text)
					fieldValue := aws.StringValue(field.ValueDetection.Text)

					switch fieldType {
					case "ITEM":
						item.Name = fieldValue
					case "PRICE":
						if price, err := parseFloat(fieldValue); err == nil {
							item.TotalPrice = price
						}
					case "QUANTITY":
						if qty, err := parseFloat(fieldValue); err == nil {
							item.Quantity = qty
						}
					case "UNIT_PRICE":
						if unitPrice, err := parseFloat(fieldValue); err == nil {
							item.UnitPrice = unitPrice
						}
					case "DESCRIPTION":
						item.Description = fieldValue
					}
				}

				// Calculate unit price if not found but quantity and total price are available
				if item.UnitPrice == 0 && item.Quantity > 0 && item.TotalPrice > 0 {
					item.UnitPrice = item.TotalPrice / item.Quantity
				}

				// Calculate total price if not found but unit price and quantity are available
				if item.TotalPrice == 0 && item.UnitPrice > 0 && item.Quantity > 0 {
					item.TotalPrice = item.UnitPrice * item.Quantity
				}

				items = append(items, item)
			}
		}
	}

	return receipt, items, nil
}

// parseDate attempts to parse a date string in various formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"Jan 2, 2006",
		"2 Jan 2006",
		"January 2, 2006",
		"2006/01/02",
	}

	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseFloat parses a string to float, handling different decimal separators and currency symbols
func parseFloat(s string) (float64, error) {
	// Remove currency symbols and non-numeric characters
	re := regexp.MustCompile(`[^0-9.,]`)
	s = re.ReplaceAllString(s, "")

	// Replace comma with dot for decimal separator
	s = strings.ReplaceAll(s, ",", ".")

	return strconv.ParseFloat(s, 64)
}

// SaveImage saves the uploaded image to the storage directory
func (s *OCRService) SaveImage(fileData []byte, fileName string) (string, error) {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate a unique filename
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(fileName)
	newFileName := fmt.Sprintf("receipt-%s%s", timestamp, ext)
	filePath := filepath.Join(s.uploadDir, newFileName)

	// Write file to disk
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return filePath, nil
}

// Note: For a production system, you'll need to integrate with a proper OCR service
// Options include:
// 1. Use the Google Cloud Vision API
// 2. Use the AWS Textract service
// 3. Use the Microsoft Azure Computer Vision API
// 4. Install and use Tesseract locally (requires more processing work)
//
// The commented code below would be used with a proper OCR implementation:

/*
// extractTextFromImage extracts text from image using an OCR service
func (s *OCRService) extractTextFromImage(imagePath string) (string, error) {
	// This would integrate with your chosen OCR service
	// For now, return a mock response
	return "Sample Store\n01/04/2023\nBread 5.99\nMilk 2 x 4.50 = 9.00\nEggs 6.99\n...", nil
}

// parseReceiptText extracts structured data from receipt OCR text
func (s *OCRService) parseReceiptText(text string) (*models.Receipt, []*models.ReceiptItem, error) {
	lines := strings.Split(text, "\n")

	// Initialize receipt with default values
	receipt := &models.Receipt{
		PurchaseDate: time.Now(), // Default to current date if not found
		StoreName:    extractStoreName(lines),
	}

	// Try to extract date
	if date, ok := extractDate(lines); ok {
		receipt.PurchaseDate = date
	}

	// Extract items and calculate total
	items := extractItems(lines)

	// Calculate total amount
	var totalAmount float64
	for _, item := range items {
		totalAmount += item.TotalPrice
	}
	receipt.TotalAmount = totalAmount

	return receipt, items, nil
}

// extractStoreName tries to find the store name in the receipt
func extractStoreName(lines []string) string {
	// Usually, the store name is at the top of the receipt
	// This is a simplistic approach and might need refinement
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return "Unknown Store"
}

// extractDate attempts to find and parse a date in the receipt
func extractDate(lines []string) (time.Time, bool) {
	datePatterns := []string{
		`(\d{2}/\d{2}/\d{4})`,                    // DD/MM/YYYY
		`(\d{2}-\d{2}-\d{4})`,                    // DD-MM-YYYY
		`(\d{2}\.\d{2}\.\d{4})`,                  // DD.MM.YYYY
		`(\d{2}/\d{2}/\d{2})`,                    // DD/MM/YY
		`(\d{1,2}\s+(?:de\s+)?[A-Za-zçÇãÃõÕêÊ]+\s+(?:de\s+)?\d{4})`, // Brazilian date format (e.g., "01 de Janeiro de 2022")
	}

	for _, line := range lines {
		for _, pattern := range datePatterns {
			re := regexp.MustCompile(pattern)
			match := re.FindStringSubmatch(line)
			if len(match) > 1 {
				// Try different date parsing formats
				formats := []string{
					"02/01/2006",
					"02-01-2006",
					"02.01.2006",
					"02/01/06",
				}

				for _, format := range formats {
					if date, err := time.Parse(format, match[1]); err == nil {
						return date, true
					}
				}

				// Try to parse Brazilian date format
				if date, success := parseBrazilianDate(match[1]); success {
					return date, true
				}
			}
		}
	}

	return time.Time{}, false
}

// parseBrazilianDate parses dates in Brazilian format (e.g., "01 de Janeiro de 2022")
func parseBrazilianDate(dateStr string) (time.Time, bool) {
	months := map[string]int{
		"janeiro": 1, "fevereiro": 2, "março": 3, "marco": 3, "abril": 4,
		"maio": 5, "junho": 6, "julho": 7, "agosto": 8,
		"setembro": 9, "outubro": 10, "novembro": 11, "dezembro": 12,
	}

	dateStr = strings.ToLower(dateStr)
	dateStr = strings.ReplaceAll(dateStr, "de", "")
	parts := strings.Fields(dateStr)

	if len(parts) < 3 {
		return time.Time{}, false
	}

	day, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, false
	}

	var month int
	for name, num := range months {
		if strings.Contains(parts[1], name) {
			month = num
			break
		}
	}
	if month == 0 {
		return time.Time{}, false
	}

	year, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return time.Time{}, false
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local), true
}

// extractItems tries to find items and their prices in the receipt
func extractItems(lines []string) []*models.ReceiptItem {
	var items []*models.ReceiptItem

	// Look for item patterns in each line
	// This is a simplified approach and might need customization
	itemPattern := regexp.MustCompile(`(.+?)\s+(\d+(?:[.,]\d+)?)\s*x\s*(\d+(?:[.,]\d+)?)\s*(?:=)?\s*(\d+(?:[.,]\d+)?)`)
	simpleItemPattern := regexp.MustCompile(`(.+?)\s+(\d+(?:[.,]\d+)?)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to match detailed item pattern (with quantity and unit price)
		matches := itemPattern.FindStringSubmatch(line)
		if len(matches) >= 5 {
			name := strings.TrimSpace(matches[1])
			quantity, _ := parseFloat(matches[2])
			unitPrice, _ := parseFloat(matches[3])
			totalPrice, _ := parseFloat(matches[4])

			items = append(items, &models.ReceiptItem{
				Name:        name,
				Quantity:    quantity,
				UnitPrice:   unitPrice,
				TotalPrice:  totalPrice,
			})
			continue
		}

		// Try to match simple item pattern (just name and price)
		matches = simpleItemPattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			name := strings.TrimSpace(matches[1])
			price, _ := parseFloat(matches[2])

			items = append(items, &models.ReceiptItem{
				Name:        name,
				Quantity:    1.0,
				UnitPrice:   price,
				TotalPrice:  price,
			})
		}
	}

	return items
}

// parseFloat parses a string to float, handling different decimal separators
func parseFloat(s string) (float64, error) {
	// Replace comma with dot for decimal separator
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}
*/
