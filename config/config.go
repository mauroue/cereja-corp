package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// Config holds all configuration for the application
type Config struct {
	Server ServerConfig `json:"server"`
	DB     DBConfig     `json:"db"`
}

// ServerConfig contains server-specific configuration
type ServerConfig struct {
	Port         string `json:"port"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
}

// DBConfig contains database configuration
type DBConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

var (
	config     *Config
	configOnce sync.Once
)

// Get returns the singleton instance of the configuration
func Get() *Config {
	configOnce.Do(func() {
		config = &Config{
			Server: ServerConfig{
				Port:         "8080",
				ReadTimeout:  10,
				WriteTimeout: 10,
			},
			DB: DBConfig{
				Host:     "postgres",
				Port:     "5432",
				Username: "postgres",
				Password: "postgres",
				Database: "cereja",
			},
		}

		// Try to load from file if exists
		if _, err := os.Stat("config.json"); err == nil {
			file, err := os.Open("config.json")
			if err != nil {
				log.Printf("Failed to open config file: %v", err)
				return
			}
			defer file.Close()

			decoder := json.NewDecoder(file)
			if err := decoder.Decode(config); err != nil {
				log.Printf("Failed to decode config file: %v", err)
			}
		}

		// Override with environment variables if they exist
		if host := os.Getenv("DB_HOST"); host != "" {
			config.DB.Host = host
		}
		if port := os.Getenv("DB_PORT"); port != "" {
			config.DB.Port = port
		}
		if user := os.Getenv("DB_USER"); user != "" {
			config.DB.Username = user
		}
		if password := os.Getenv("DB_PASSWORD"); password != "" {
			config.DB.Password = password
		}
		if name := os.Getenv("DB_NAME"); name != "" {
			config.DB.Database = name
		}
	})

	return config
}

// Save persists the current configuration to a file
func Save() error {
	file, err := os.Create("config.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(Get())
}
