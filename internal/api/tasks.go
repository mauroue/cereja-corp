package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mauroue/cereja-corp/internal/models"
)

// Task represents a task in the system
type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// In-memory storage for tasks - in a real application, use a database
var tasks = []Task{}

// SetupTaskRoutes configures the routes for task management
func SetupTaskRoutes(router *gin.RouterGroup) {
	taskRoutes := router.Group("/tasks")
	{
		taskRoutes.GET("", getAllTasks)
		taskRoutes.GET("/:id", getTaskByID)
		taskRoutes.POST("", createTask)
		taskRoutes.PUT("/:id", updateTask)
		taskRoutes.DELETE("/:id", deleteTask)
	}
}

// getAllTasks returns all tasks
func getAllTasks(c *gin.Context) {
	c.JSON(http.StatusOK, tasks)
}

// getTaskByID returns a specific task by ID
func getTaskByID(c *gin.Context) {
	id := c.Param("id")

	for _, task := range tasks {
		if task.ID == id {
			c.JSON(http.StatusOK, task)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
}

// createTask creates a new task
func createTask(c *gin.Context) {
	var newTask Task
	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a simple ID - in a real app, use UUID or similar
	newTask.ID = models.GenerateID()
	tasks = append(tasks, newTask)

	c.JSON(http.StatusCreated, newTask)
}

// updateTask updates an existing task
func updateTask(c *gin.Context) {
	id := c.Param("id")
	var updatedTask Task

	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, task := range tasks {
		if task.ID == id {
			updatedTask.ID = id
			tasks[i] = updatedTask
			c.JSON(http.StatusOK, updatedTask)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
}

// deleteTask deletes a task by ID
func deleteTask(c *gin.Context) {
	id := c.Param("id")

	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
}
