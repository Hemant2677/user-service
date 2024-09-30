package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"user-service/database"
	"user-service/models"
	"user-service/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Create a new user
func NewUser(c *gin.Context) {
	var user models.User
	var err error
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// check if any filed users sending is empty
	if user.Name == "" || user.MobileNumber == "" || user.Email == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	// Check if the email already exists in the database
	row := database.Db.QueryRow("SELECT id FROM users WHERE email=$1", user.Email)
	var id int
	err = row.Scan(&id)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// check if mobile number is already in the database
	row = database.Db.QueryRow("SELECT id FROM users WHERE mobile_number=$1", user.MobileNumber)
	err = row.Scan(&id)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Mobile number already registered,try using other mobile number"})
		return
	}

	// Validate the age range
	if user.Age < 18 || user.Age > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Age should be between 18 and 100"})
		return
	}

	// Validate the age range
	if user.Age < 18 || user.Age > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Age should be between 18 and 100"})
		return
	}

	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Store the hashed password in the database
	sqlStatement := `INSERT INTO users (name, age, mobile_number, email, password) VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	err = database.Db.QueryRow(sqlStatement, user.Name, user.Age, user.MobileNumber, user.Email, hashedPassword).Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("New user created:", user)
	c.JSON(http.StatusCreated, user)
}

func Login(c *gin.Context) {
	var user models.User
	var err error
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	row := database.Db.QueryRow("SELECT * FROM users WHERE email=$1", user.Email)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, map[string]any{"error": "Invalid email or password"})
		return
	}

	var dbUser models.User
	if err := row.Scan(&dbUser.ID, &dbUser.Name, &dbUser.Age, &dbUser.MobileNumber, &dbUser.Email, &dbUser.Password); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, map[string]any{"error": "Invalid email or password"})
		} else {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		}
		return
	}

	// Compare the hashed password with the one stored in the database
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, map[string]any{"error": "Invalid email or password"})
		return
	}

	// Generate a JWT token for the user
	tokenString, err := utils.GenerateJWT(dbUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK,
		map[string]any{
			"status": "successfully logged in",
			"token":  tokenString})
}
