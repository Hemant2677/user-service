package handlers

import (
	"log"
	"net/http"
	"strconv"

	"user-service/internal/database"

	"user-service/pkg/utils"

	"github.com/gin-gonic/gin"
)

type UserResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Age          int    `json:"age"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
}

func CreateUserHandler(c *gin.Context) {
	var userRequest struct {
		Name         string `json:"name"`
		Age          int    `json:"age"`
		MobileNumber string `json:"mobile_number"`
		Email        string `json:"email"`
		Password     string `json:"password"`
	}

	// Bind JSON request to user struct
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid input", "details": err.Error()})
		return
	}

	

	// Call the database function to create a new user
	userResponse, err := database.Createnewuser(&database.User{
		Name:         userRequest.Name,
		Age:          userRequest.Age,
		MobileNumber: userRequest.MobileNumber,
		Email:        userRequest.Email,
		Password:     userRequest.Password,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	// Map the user data to UserResponse to exclude the password
	response := UserResponse{
		ID:           userResponse.ID,
		Name:         userResponse.Name,
		Age:          userResponse.Age,
		MobileNumber: userResponse.MobileNumber,
		Email:        userResponse.Email,
	}


	// Return the created user as JSON
	c.JSON(http.StatusCreated,
		map[string]any{"status": "sucesfull", "user": response})
}

func LoginHandler(c *gin.Context) {
	var userRequest struct {
		Name         string `json:"name"`
		Age          int    `json:"age"`
		MobileNumber string `json:"mobile_number"`
		Email        string `json:"email"`
		Password     string `json:"password"`
	}

	// Bind JSON request to user struct
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Generate a JWT token for the authenticated user
	token, err := utils.GenerateJWT(userRequest.Email)
	if err != nil {
		log.Printf("Error generating JWT token: %v", err)
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Internal server error"})
		return
	}

	// Return the JWT token
	c.JSON(http.StatusOK, map[string]any{"token": token})
}

func GetAllUsersHandler(c *gin.Context) {
	// Parse query parameters for pagination
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid page number"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid limit number"})
		return
	}

	// Call the database function to get users
	users, total, err := database.Getallusers(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	// Set headers for pagination and CORS
	headers := c.Writer.Header()
	headers.Set("X-Total-Count", strconv.Itoa(total))
	headers.Set("Access-Control-Expose-Headers", "X-Total-Count")
	headers.Set("Access-Control-Allow-Headers", "X-Total-Count")
	headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	headers.Set("Access-Control-Allow-Origin", "*")
	headers.Set("Content-Type", "application/json")
	headers.Set("X-RateLimit-Limit", "100")
	headers.Set("X-RateLimit-Remaining", strconv.Itoa(100))

	// Return users and metadata as JSON
	c.JSON(http.StatusOK, map[string]any{
		"users":       users,
		"total_users": total,
		"page":        page,
		"limit":       limit,
	})
}

func GetUserByIDHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "User ID is required"})
		return
	}
	// Call the database function to fetch the user by ID
	user, err := database.Getuserbyid(id)
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}

	// Return the user as JSON
	c.JSON(http.StatusOK, map[string]any{"user": user})
}
