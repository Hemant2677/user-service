package routes

import (
	"user-service/internal/handlers"
	"user-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Public routes
	r.POST("/login", handlers.Login)
	r.POST("/register", handlers.CreateUserHandler)

	authGroup := r.Group("/")
	authGroup.Use(middleware.AuthMiddleware())

	authGroup.GET("/users", handlers.GetAllUsersHandler)
	authGroup.GET("/users/:id", handlers.GetUserByIDHandler)
	authGroup.PUT("/users/:id", handlers.UpdateUser)
}
