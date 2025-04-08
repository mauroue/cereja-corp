package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/mauroue/cereja-corp/config"
)

var (
	db     *sql.DB
	dbOnce sync.Once
)

// Connect establishes a connection to the database
func Connect() (*sql.DB, error) {
	var err error

	dbOnce.Do(func() {
		cfg := config.Get()
		connStr := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password, cfg.DB.Database,
		)

		// Log the connection string for debugging
		log.Printf("Connecting to database with: host=%s port=%s dbname=%s", cfg.DB.Host, cfg.DB.Port, cfg.DB.Database)

		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return
		}

		err = db.Ping()
	})

	return db, err
}

// GetDB returns the singleton database connection
func GetDB() (*sql.DB, error) {
	if db == nil {
		return Connect()
	}
	return db, nil
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
