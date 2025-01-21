package routes

import (
	"gin-mongo-api/controllers"
	"gin-mongo-api/middlewares"

	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.Engine) {

	router.POST("/register", controllers.Register())
	router.POST("/login", controllers.Login())

	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	{
		protected.POST("/user", controllers.CreateUser())
		protected.GET("/user/:userId", controllers.GetAUser())
		protected.PUT("/user/:userId", controllers.EditAUser())
		protected.DELETE("/user/:userId", controllers.DeleteAUser())
		protected.GET("/users", controllers.GetAllUsers())
	}
}
