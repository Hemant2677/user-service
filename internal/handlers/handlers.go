package handlers

import (
	"net/http"
	"strconv"

	"user-service/internal/database"
	"user-service/internal/models"

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
	user := models.User{
		Name:         userRequest.Name,
		Age:          userRequest.Age,
		MobileNumber: userRequest.MobileNumber,
		Email:        userRequest.Email,
		Password:     userRequest.Password,
	}
	userResponse, err := database.Createnewuser(&user)
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
		map[string]any{"status": "successful", "user": response})
}

func Login(c *gin.Context) {
	var loginRequest database.User
	//      var err error

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid Request"})
		return
	}

	// Check if empty field
	if loginRequest.Email == "" || loginRequest.Password == "" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Email and password are required"})
		return
	}

	// Fetch the full user details, including the hashed password
	user, err := database.FetchUserByEmail(loginRequest.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}

	// Compare the provided password with the hashed password
	err = utils.ComparePasswords(user.Password, loginRequest.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]any{"error": "invalid password"})
		return
	}

	// Generate the JWT Token
	token, err := utils.GenerateJWT(user.ID, user.Name, user.Email)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"error": "Failed to generate JWT token"},
		)
		return
	}

	// Send the token as response
	c.JSON(http.StatusOK, map[string]any{
		"message": "Successfully logged in",
		"token":   token,
		"status":  "success",
	})

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

// UpdateUser is the handler for updating user details
func UpdateUser(c *gin.Context) {
	// Extract token from the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	// Split the header to extract the token part
	tokenString := authHeader

	// Extract user info (ID, Name, Email) from token
	userID, _, _, err := utils.ExtractUserInfo(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Get the ID of the user being updated from the request path
	paramID := c.Param("id") // Assuming the user ID is in the URL path, e.g., /users/:id
	if paramID != strconv.Itoa(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own information"})
		return
	}

	// Extract new details from the request body
	var userRequest struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Proceed to update the user information in the database...

	err = database.UpdateUserByID(userID, userRequest.Name, userRequest.Age)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}
