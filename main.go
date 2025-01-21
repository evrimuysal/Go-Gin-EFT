package main

import (
	"gin-mongo-api/configs"
	"gin-mongo-api/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

// Error message as a constant
const ErrPortNotSet = "$PORT must be set"

func main() {
	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		log.Fatal(ErrPortNotSet)
	}

	// Initialize router
	router := setupRouter()

	// Start the server
	router.Run(":" + portEnv)
}

// setupRouter sets up the Gin router and initializes configurations and routes
func setupRouter() *gin.Engine {
	router := gin.Default()

	// Connect to the database
	configs.ConnectDB()

	// Apply middleware
	router.Use(gin.Logger())

	// Register routes
	routes.UserRoute(router)

	return router
}
