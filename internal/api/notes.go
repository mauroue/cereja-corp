package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mauroue/cereja-corp/internal/models"
)

// Note represents a note in the system
type Note struct {
	ID      string   `json:"id"`
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

// In-memory storage for notes - in a real application, use a database
var notes = []Note{}

// SetupNoteRoutes configures the routes for note management
func SetupNoteRoutes(router *gin.RouterGroup) {
	noteRoutes := router.Group("/notes")
	{
		noteRoutes.GET("", getAllNotes)
		noteRoutes.GET("/:id", getNoteByID)
		noteRoutes.POST("", createNote)
		noteRoutes.PUT("/:id", updateNote)
		noteRoutes.DELETE("/:id", deleteNote)
	}
}

// getAllNotes returns all notes
func getAllNotes(c *gin.Context) {
	c.JSON(http.StatusOK, notes)
}

// getNoteByID returns a specific note by ID
func getNoteByID(c *gin.Context) {
	id := c.Param("id")

	for _, note := range notes {
		if note.ID == id {
			c.JSON(http.StatusOK, note)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
}

// createNote creates a new note
func createNote(c *gin.Context) {
	var newNote Note
	if err := c.ShouldBindJSON(&newNote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a simple ID - in a real app, use UUID or similar
	newNote.ID = models.GenerateID()
	notes = append(notes, newNote)

	c.JSON(http.StatusCreated, newNote)
}

// updateNote updates an existing note
func updateNote(c *gin.Context) {
	id := c.Param("id")
	var updatedNote Note

	if err := c.ShouldBindJSON(&updatedNote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, note := range notes {
		if note.ID == id {
			updatedNote.ID = id
			notes[i] = updatedNote
			c.JSON(http.StatusOK, updatedNote)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
}

// deleteNote deletes a note by ID
func deleteNote(c *gin.Context) {
	id := c.Param("id")

	for i, note := range notes {
		if note.ID == id {
			notes = append(notes[:i], notes[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Note deleted"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
}
