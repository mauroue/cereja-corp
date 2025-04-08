# Cereja Corp

A monolithic Go application using Gin framework to provide various personal services and tools.

## Project Structure

```
cereja-corp/
├── cmd/          # Main applications
├── internal/     # Private application code
│   ├── api/      # API handlers
│   ├── db/       # Database layer
│   ├── models/   # Data models
│   └── server/   # Server configuration
├── pkg/          # Public libraries
├── api/          # API definitions
├── config/       # Configuration
├── scripts/      # Build and deployment scripts
└── docs/         # Documentation
```

## Features

- RESTful API using Gin framework
- Modular architecture for easy extension
- PostgreSQL database integration
- Task management API
- Notes management API
- Docker and docker-compose support

## Getting Started

### Prerequisites

- Go 1.16+
- PostgreSQL (optional, can use Docker)
- AWS account (for Textract OCR functionality)

### Installation

1. Clone the repository
2. Install dependencies:
   ```
   go mod tidy
   ```
3. Set up AWS credentials (required for Textract OCR):
   - Set the AWS credentials using environment variables:
     ```
     export AWS_ACCESS_KEY_ID=your_access_key_id
     export AWS_SECRET_ACCESS_KEY=your_secret_access_key
     export AWS_REGION=your_preferred_region
     ```
   - Or configure AWS credentials using the AWS CLI:
     ```
     aws configure
     ```
4. Run the application:
   ```
   go run cmd/main.go
   ```

The server will start on port 8080.

## AWS Textract Integration

The application uses AWS Textract for optical character recognition (OCR) of receipts. This allows for automatic extraction of:

- Store name
- Purchase date
- Total amount
- Individual items and prices

### Configuration

Configure AWS for Textract using environment variables:
- `AWS_REGION`: AWS region where Textract service will be used (default: "us-east-1")
- `AWS_ACCESS_KEY_ID`: Your AWS access key
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret key

If AWS credentials are not provided or Textract is unavailable, the application will fall back to using mock data for testing purposes.

### Permissions

Ensure your AWS user has the following permissions:
- `textract:AnalyzeExpense`

## API Endpoints

### Tasks API

- `GET /api/v1/tasks` - List all tasks
- `GET /api/v1/tasks/:id` - Get a specific task
- `POST /api/v1/tasks` - Create a new task
- `PUT /api/v1/tasks/:id` - Update a task
- `DELETE /api/v1/tasks/:id` - Delete a task

### Notes API

- `GET /api/v1/notes` - List all notes  
- `GET /api/v1/notes/:id` - Get a specific note
- `POST /api/v1/notes` - Create a new note
- `PUT /api/v1/notes/:id` - Update a note
- `DELETE /api/v1/notes/:id` - Delete a note

### Receipts API

- `POST /receipts/upload` - Upload and process a receipt image
- `GET /receipts/:id` - Get a specific receipt
- `GET /receipts/:id/items` - Get items for a specific receipt
- `GET /receipts` - List all receipts

## Development

### Running with Docker

```
docker-compose up -d
```

### Adding New Features

To add new services, follow these steps:

1. Create handlers in the appropriate package in `internal/api/`
2. Register routes in the server package
3. Add appropriate tests

## License

This project is licensed under the MIT License.