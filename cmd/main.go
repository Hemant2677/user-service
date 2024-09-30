package main
import (
	"user-service/database"
	"user-service/internal/handlers"
	"user-service/middleware"
	"github.com/gin-gonic/gin"
)
func main() {
	r := gin.Default()
	r.POST("/register", handlers.NewUser)
	r.POST("/login", handlers.Login)
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/users", database.Getallusers)
		auth.GET("/users/:id", database.Byid)
	}
	r.Run(":3000")
}
