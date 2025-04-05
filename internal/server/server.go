package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mauroue/cereja-corp/internal/api"
)

// Server encapsulates the Gin router and other dependencies
type Server struct {
	Router *gin.Engine
}

// NewServer creates a new server instance
func NewServer(router *gin.Engine) *Server {
	return &Server{
		Router: router,
	}
}

// SetupRoutes configures all the routes for the application
func (s *Server) SetupRoutes() {
	// Setup middleware
	s.Router.Use(gin.Logger())
	s.Router.Use(gin.Recovery())

	// Health check endpoint
	s.Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API routes
	apiGroup := s.Router.Group("/api/v1")
	{
		// Setup different service endpoints
		api.SetupTaskRoutes(apiGroup)
		api.SetupNoteRoutes(apiGroup)
		// Add more service routes as needed
	}
}
