package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jnp/go-structured-logs/pkg/httpclient"
	"github.com/jnp/go-structured-logs/pkg/logger"
	"github.com/jnp/go-structured-logs/pkg/middleware"
)

func main() {
	// Create logger configuration
	config := logger.DefaultConfig()
	config.ApplicationName = "demo-service"

	// Configure defaults for simple usage
	logger.ConfigureDefaults(config)

	// Create logger
	log := logger.New("main")

	// Create Gin router
	r := gin.New()
	r.Use(gin.Recovery())

	// Add logging middleware
	r.Use(middleware.LoggingMiddleware(log, config))

	// Example routes
	r.GET("/api/users", getUsers)
	r.GET("/api/users/:id", getUser)
	r.POST("/api/users", createUser)
	r.GET("/api/external", callExternalAPI)
	r.GET("/health", healthCheck)

	// Start server
	log.Info(context.Background(), "Starting server on :8080")
	r.Run(":8080")
}

func getUsers(c *gin.Context) {
	log := logger.New("UserHandler")
	ctx := c.Request.Context()

	log.Info(ctx, "Fetching all users")

	users := []map[string]interface{}{
		{"id": "1", "name": "John Doe", "email": "john@example.com"},
		{"id": "2", "name": "Jane Doe", "email": "jane@example.com"},
	}

	log.Debug(ctx, "Found users", map[string]interface{}{"count": len(users)})

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   users,
	})
}

func getUser(c *gin.Context) {
	log := logger.New("UserHandler")
	ctx := c.Request.Context()

	id := c.Param("id")
	log.Info(ctx, "Fetching user", map[string]interface{}{"user_id": id})

	user := map[string]interface{}{
		"id":    id,
		"name":  "John Doe",
		"email": "john@example.com",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   user,
	})
}

func createUser(c *gin.Context) {
	log := logger.New("UserHandler")
	ctx := c.Request.Context()

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		log.WarnError(ctx, "Invalid request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	log.Info(ctx, "Creating user", map[string]interface{}{"email": input["email"]})

	// Simulate user creation
	input["id"] = "123"

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   input,
	})
}

func callExternalAPI(c *gin.Context) {
	log := logger.New("ExternalAPIHandler")
	ctx := c.Request.Context()
	config := logger.GetDefaultConfig()

	log.Info(ctx, "Calling external API")

	// Create HTTP client with logging
	client := httpclient.NewClient(log, config)

	// Make external request
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://jsonplaceholder.typicode.com/todos/1", nil)
	resp, err := client.Do(req)
	if err != nil {
		log.ErrorWithErr(ctx, "External API call failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to call external API",
		})
		return
	}
	defer resp.Body.Close()

	log.Info(ctx, "External API call successful", map[string]interface{}{
		"status_code": resp.StatusCode,
	})

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "External API called successfully",
	})
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
