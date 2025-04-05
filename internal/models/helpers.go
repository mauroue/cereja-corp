package models

import (
	"fmt"
	"math/rand"
	"time"
)

// Initialize the random seed
func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateID creates a simple random ID
// In a production environment, consider using UUID or similar
func GenerateID() string {
	return fmt.Sprintf("%d", rand.Int63())
}
