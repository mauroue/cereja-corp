package receipts

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// WebHandler manages HTTP requests for receipt web interface
type WebHandler struct {
	api  *Handler
	repo *Repository
}

// NewWebHandler creates a new web handler for receipts
func NewWebHandler(api *Handler, templatesDir string) (*WebHandler, error) {
	// Get repository from API handler
	repo := api.repo

	return &WebHandler{
		api:  api,
		repo: repo,
	}, nil
}

// RegisterRoutes registers the web handler routes
func (h *WebHandler) RegisterRoutes(router *gin.Engine) {
	// Serve static files
	router.Static("/static", "./internal/receipts/static")

	// Web routes
	web := router.Group("/receipts-web")
	{
		web.GET("/", h.HomePage)
		web.GET("/upload", h.UploadPage)
		web.GET("/list", h.ListPage)
		web.GET("/view/:id", h.ViewPage)

		// HTMX endpoints
		web.POST("/htmx/upload", h.HtmxUpload)
		web.GET("/htmx/receipts", h.HtmxListReceipts)
		web.GET("/htmx/receipt/:id", h.HtmxGetReceipt)
		web.GET("/htmx/receipt/:id/items", h.HtmxGetReceiptItems)
	}
}

// Common HTML layout handling
func renderPageWithLayout(title string, content string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Receipt Scanner</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body>
    <header>
        <div class="container navbar">
            <div class="logo">Receipt Scanner</div>
            <nav class="nav-links">
                <a href="/receipts-web/">Home</a>
                <a href="/receipts-web/upload">Upload</a>
                <a href="/receipts-web/list">My Receipts</a>
            </nav>
        </div>
    </header>

    <main>
        <div class="container">
            %s
        </div>
    </main>

    <footer>
        <div class="container text-center">
            <p>&copy; %d Receipt Scanner App</p>
        </div>
    </footer>
</body>
</html>
`, title, content, time.Now().Year())
}

// HomePage renders the home page
func (h *WebHandler) HomePage(c *gin.Context) {
	content := `
<div class="card">
    <div class="card-header">
        <h1 class="card-title">Receipt Scanner App</h1>
    </div>
    <p>Welcome to the Receipt Scanner App! This application helps you track your purchases by scanning and storing your receipts.</p>
    <p>With our app, you can:</p>
    <ul>
        <li>Upload images of your receipts</li>
        <li>Automatically extract store, item, and price information</li>
        <li>Keep track of all your purchases in one place</li>
        <li>View detailed reports of your spending</li>
    </ul>
    <div class="mt-4">
        <a href="/receipts-web/upload" class="btn btn-primary">Upload a Receipt</a>
        <a href="/receipts-web/list" class="btn btn-secondary">View My Receipts</a>
    </div>
</div>
`
	html := renderPageWithLayout("Home", content)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// UploadPage renders the upload page
func (h *WebHandler) UploadPage(c *gin.Context) {
	content := `
<div class="card">
    <div class="card-header">
        <h1 class="card-title">Upload Receipt</h1>
    </div>
    
    <!-- Error messages will be displayed here -->
    <div id="upload-error-container"></div>
    
    <form hx-post="/receipts-web/htmx/upload" 
          hx-encoding="multipart/form-data" 
          hx-indicator="#form-submit-indicator"
          hx-target="#upload-error-container"
          hx-swap="innerHTML"
          hx-validate="true"
          class="upload-form">
        
        <div class="form-group">
            <label for="receipt">Receipt Image</label>
            <div class="file-upload">
                <label for="receipt">
                    <div class="file-upload-icon">📷</div>
                    <div class="file-upload-text" id="file-upload-text">Click to select a receipt image or drag and drop</div>
                </label>
                <input type="file" id="receipt" name="receipt" accept="image/*" required
                       onchange="updateFileName(this)">
            </div>
            <div id="file-selected" class="file-selected-info"></div>
        </div>
        
        <div class="form-group text-center">
            <button type="submit" class="btn btn-primary" id="upload-button" onclick="validateUpload(event)">
                <span id="form-submit-indicator" class="htmx-indicator">
                    <span class="loading-spinner"></span> Processing...
                </span>
                <span class="htmx-indicator-inverse">Upload Receipt</span>
            </button>
        </div>
    </form>
</div>

<script>
    // Display the selected filename when a file is chosen
    function updateFileName(input) {
        const fileSelectedDiv = document.getElementById('file-selected');
        const fileUploadText = document.getElementById('file-upload-text');
        
        if (input.files && input.files[0]) {
            const fileName = input.files[0].name;
            const fileSize = (input.files[0].size / 1024).toFixed(2) + ' KB';
            
            fileSelectedDiv.innerHTML = '<div class="alert alert-info">' +
                '<strong>File selected:</strong> ' + fileName + ' (' + fileSize + ')' +
                '</div>';
            fileUploadText.textContent = "Change file";
            
            // Clear any previous error message
            const errorContainer = document.getElementById('upload-error-container');
            errorContainer.innerHTML = '';
        } else {
            fileSelectedDiv.innerHTML = '';
            fileUploadText.textContent = "Click to select a receipt image or drag and drop";
        }
    }
    
    // Validate that a file is selected before submitting
    function validateUpload(event) {
        const fileInput = document.getElementById('receipt');
        const errorContainer = document.getElementById('upload-error-container');
        
        if (!fileInput.files || fileInput.files.length === 0) {
            event.preventDefault();
            errorContainer.innerHTML = '<div class="alert alert-danger">' +
                '<strong>Error:</strong> Please select a file to upload' +
                '</div>';
            return false;
        }
        return true;
    }
</script>
`
	html := renderPageWithLayout("Upload Receipt", content)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// ListPage renders the list page
func (h *WebHandler) ListPage(c *gin.Context) {
	content := `
<div class="card">
    <div class="card-header">
        <h1 class="card-title">My Receipts</h1>
    </div>
    
    <div id="receipts-list"
         hx-get="/receipts-web/htmx/receipts"
         hx-trigger="load"
         hx-indicator="#receipts-loading">
        <div class="text-center mt-3">
            <div id="receipts-loading" class="loading-spinner htmx-indicator"></div>
            <p>Loading receipts...</p>
        </div>
    </div>
</div>
`
	html := renderPageWithLayout("My Receipts", content)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// ViewPage renders the view page for a receipt
func (h *WebHandler) ViewPage(c *gin.Context) {
	idStr := c.Param("id")
	content := fmt.Sprintf(`
<div class="card">
    <div class="card-header">
        <h1 class="card-title">View Receipt</h1>
        <a href="/receipts-web/list" class="btn btn-secondary">Back to List</a>
    </div>
    
    <div id="receipt-details"
         hx-get="/receipts-web/htmx/receipt/%s"
         hx-trigger="load"
         hx-indicator="#details-loading">
        <div class="text-center mt-3">
            <div id="details-loading" class="loading-spinner htmx-indicator"></div>
            <p>Loading receipt details...</p>
        </div>
    </div>
    
    <div id="receipt-items"
         hx-get="/receipts-web/htmx/receipt/%s/items"
         hx-trigger="load"
         hx-indicator="#items-loading">
        <div class="text-center mt-3">
            <div id="items-loading" class="loading-spinner htmx-indicator"></div>
            <p>Loading receipt items...</p>
        </div>
    </div>
</div>
`, idStr, idStr)

	html := renderPageWithLayout("View Receipt", content)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// Utility function to create consistent error responses
func createErrorResponse(message string) string {
	return fmt.Sprintf(`
	<div class="alert alert-danger">
		<strong>Error:</strong> %s
	</div>
	`, message)
}

// Utility function to create success responses
func createSuccessResponse(message string) string {
	return fmt.Sprintf(`
	<div class="alert alert-success">
		<strong>Success:</strong> %s
	</div>
	`, message)
}

// HtmxUpload handles receipt upload via HTMX
func (h *WebHandler) HtmxUpload(c *gin.Context) {
	// Parse multipart form with 32MB max memory
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("Failed to parse form")))
		return
	}

	// Get the uploaded file
	file, header, err := c.Request.FormFile("receipt")
	if err != nil {
		c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("No file uploaded. Please select a receipt image.")))
		return
	}
	defer file.Close()

	// Validate file size (max 10MB)
	if header.Size > 10*1024*1024 {
		c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("File is too large. Maximum file size is 10MB.")))
		return
	}

	// Validate file type
	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	validExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".pdf": true}
	if !validExts[fileExt] {
		c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("Invalid file type. Please upload an image file (jpg, png, gif, bmp) or PDF.")))
		return
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("Failed to read file")))
		return
	}

	// Save the image
	imagePath, err := h.api.ocrService.SaveImage(fileData, header.Filename)
	if err != nil {
		c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("Failed to save image: "+err.Error())))
		return
	}

	// Process the receipt
	receipt, items, err := h.api.ocrService.ProcessReceipt(imagePath)
	if err != nil {
		errorMsg := "Failed to process receipt"
		if strings.Contains(err.Error(), "AWS Textract client not available") {
			errorMsg = `
			<div class="alert alert-danger">
				<strong>AWS credentials not configured</strong>
				<p>Please set the following environment variables to use the receipt scanner:</p>
				<ul>
					<li>AWS_ACCESS_KEY_ID - Your AWS access key</li>
					<li>AWS_SECRET_ACCESS_KEY - Your AWS secret key</li>
					<li>AWS_REGION - AWS region (e.g., us-east-1)</li>
				</ul>
				<p>These credentials are required to use AWS Textract for receipt processing.</p>
			</div>
			`
		} else {
			errorMsg = createErrorResponse("Error processing receipt: " + err.Error())
		}

		c.Data(http.StatusOK, "text/html", []byte(errorMsg))

		// Remove the saved image if processing failed
		if imagePath != "" {
			os.Remove(imagePath)
		}

		return
	}

	// Ensure the default store exists
	if err := h.api.repo.EnsureDefaultStore(); err != nil {
		c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("Failed to ensure store exists: "+err.Error())))
		return
	}

	// Save receipt data to database
	receiptID, err := h.api.repo.CreateReceipt(receipt)
	if err != nil {
		c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("Failed to save receipt: "+err.Error())))
		return
	}

	// Save receipt items
	for _, item := range items {
		item.ReceiptID = receiptID
		_, err := h.api.repo.CreateReceiptItem(item)
		if err != nil {
			c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("Failed to save receipt items: "+err.Error())))
			return
		}
	}

	// Return success and redirect
	c.Header("HX-Redirect", "/receipts-web/list")
}

// HtmxListReceipts returns a list of receipts for HTMX
func (h *WebHandler) HtmxListReceipts(c *gin.Context) {
	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize := 10
	search := c.Query("search")

	// Try to get actual data from the database
	receipts, err := h.repo.ListReceipts(page, pageSize, search)

	// If we get an error or no receipts found, return a message
	if err != nil || len(receipts) == 0 {
		c.Data(http.StatusOK, "text/html", []byte(`
		<p>No receipts found. <a href="/receipts-web/upload">Upload your first receipt</a>.</p>
		`))
		return
	}

	// Get total count for pagination
	total, err := h.repo.GetReceiptsCount(search)
	if err != nil {
		total = 0
	}

	// Format receipt data and build HTML
	var html strings.Builder
	html.WriteString(`<div class="table-responsive"><table class="table"><thead><tr><th>Store</th><th>Date</th><th>Amount</th><th>Actions</th></tr></thead><tbody>`)

	for _, receipt := range receipts {
		formattedDate := formatDate(receipt.PurchaseDate)
		formattedAmount := formatCurrency(receipt.TotalAmount)

		html.WriteString(fmt.Sprintf(`
		<tr>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>
				<a href="/receipts-web/view/%d" class="btn btn-sm btn-info">View</a>
			</td>
		</tr>
		`, receipt.StoreName, formattedDate, formattedAmount, receipt.ID))
	}

	html.WriteString(`</tbody></table></div>`)

	// Add pagination if needed
	hasMore := total > page*pageSize
	if hasMore {
		nextPage := page + 1
		html.WriteString(fmt.Sprintf(`
		<div class="mt-3 text-center">
			<button class="btn btn-secondary" 
					hx-get="/receipts-web/htmx/receipts?page=%d" 
					hx-target="#receipts-list" 
					hx-swap="outerHTML">
				Load More
			</button>
		</div>
		`, nextPage))
	}

	c.Data(http.StatusOK, "text/html", []byte(html.String()))
}

// HtmxGetReceipt returns a single receipt for HTMX
func (h *WebHandler) HtmxGetReceipt(c *gin.Context) {
	// Get the receipt ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("Invalid receipt ID")))
		return
	}

	// Try to get the receipt from the database
	receipt, err := h.repo.GetReceiptByID(id)
	if err != nil {
		c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("Receipt not found")))
		return
	}

	// Format the data and build HTML
	formattedDate := formatDate(receipt.PurchaseDate)
	formattedAmount := formatCurrency(receipt.TotalAmount)

	html := fmt.Sprintf(`
	<div class="receipt-details">
		<h2>Receipt Details</h2>
		<dl class="receipt-info">
			<dt>Store:</dt>
			<dd>%s</dd>
			
			<dt>Date:</dt>
			<dd>%s</dd>
			
			<dt>Total Amount:</dt>
			<dd>%s</dd>
		</dl>
		
		<div class="receipt-image-container">
			<img src="/uploads/receipts/%s" alt="Receipt Image" class="receipt-image" />
		</div>
	</div>
	`,
		receipt.StoreName,
		formattedDate,
		formattedAmount,
		filepath.Base(receipt.ImagePath))

	c.Data(http.StatusOK, "text/html", []byte(html))
}

// HtmxGetReceiptItems returns items for a receipt for HTMX
func (h *WebHandler) HtmxGetReceiptItems(c *gin.Context) {
	// Get the receipt ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("Invalid receipt ID")))
		return
	}

	// Try to get the items from the database
	items, err := h.repo.GetReceiptItems(id)
	if err != nil || len(items) == 0 {
		c.Data(http.StatusOK, "text/html", []byte(createErrorResponse("No items found for this receipt")))
		return
	}

	// Calculate total and build HTML
	var total float64
	var html strings.Builder

	html.WriteString(`
	<div class="receipt-items">
		<h2>Receipt Items</h2>
		<div class="table-responsive">
			<table class="table">
				<thead>
					<tr>
						<th>Item</th>
						<th>Description</th>
						<th>Quantity</th>
						<th>Unit Price</th>
						<th>Total</th>
					</tr>
				</thead>
				<tbody>
	`)

	for _, item := range items {
		total += item.TotalPrice
		unitPrice := formatCurrency(item.UnitPrice)
		totalPrice := formatCurrency(item.TotalPrice)

		html.WriteString(fmt.Sprintf(`
		<tr>
			<td>%s</td>
			<td>%s</td>
			<td>%.2f</td>
			<td>%s</td>
			<td>%s</td>
		</tr>
		`, item.Name, item.Description, item.Quantity, unitPrice, totalPrice))
	}

	formattedTotal := formatCurrency(total)
	html.WriteString(fmt.Sprintf(`
				</tbody>
				<tfoot>
					<tr>
						<th colspan="4" class="text-right">Total:</th>
						<th>%s</th>
					</tr>
				</tfoot>
			</table>
		</div>
	</div>
	`, formattedTotal))

	c.Data(http.StatusOK, "text/html", []byte(html.String()))
}

// Helper functions

// formatDate formats a time.Time to a human-readable string
func formatDate(t time.Time) string {
	return t.Format("January 2, 2006")
}

// formatCurrency formats a float as a currency string
func formatCurrency(amount float64) string {
	return "$" + strconv.FormatFloat(amount, 'f', 2, 64)
}
