# Receipt Scanner App

This module provides functionality to scan purchase receipts (nota fiscal), parse the items and prices, and store the data in a database.

## Features

- Upload and store receipt images
- Extract data from receipt images (store, items, prices, date)
- Store receipt data in a structured database
- API endpoints to manage and query receipt data

## Setup

1. Ensure the database is running with the correct schema (run the migrations in `migrations/`)
2. Make sure the upload directory exists and is writable
3. The app is automatically integrated with the main application

## API Endpoints

- `POST /receipts/upload` - Upload a receipt image for processing
- `GET /receipts/:id` - Get details of a specific receipt
- `GET /receipts/:id/items` - Get all items for a specific receipt
- `GET /receipts` - List all receipts (with pagination)

## OCR Integration

The current implementation uses a placeholder for OCR functionality. For production use, you should integrate with a proper OCR service:

1. Google Cloud Vision API
2. AWS Textract
3. Microsoft Azure Computer Vision API
4. Local Tesseract installation

## Database Schema

### Receipts Table
- `id` - Primary key
- `store_id` - Reference to the store
- `store_name` - Name of the store
- `purchase_date` - Date of the purchase
- `total_amount` - Total amount of the purchase
- `image_path` - Path to the stored receipt image
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp

### Receipt Items Table
- `id` - Primary key
- `receipt_id` - Reference to the receipt
- `name` - Name of the item
- `description` - Description of the item
- `quantity` - Quantity of the item
- `unit_price` - Price per unit
- `total_price` - Total price for this item
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp

### Stores Table
- `id` - Primary key
- `name` - Store name
- `address` - Store address
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp 