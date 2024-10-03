package main

import (

	"user-service/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Gin router
	r := gin.Default()

	// Setup application routes
	routes.SetupRoutes(r)

	err:= r.Run(":3000");
	if err!=nil{
        panic(err)
    }
}
