package database

import (
	"database/sql"
	"errors"
	"fmt"
	"user-service/internal/models"
	"user-service/pkg/utils"
)

type UserResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Age          int    `json:"age"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
}

type UserRequest struct {
	Name         string `json:"name"`
	Age          int    `json:"age"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
	Password     string `json:"password"`
}

type User struct {
	ID           int
	Name         string `json:"name"`
	Age          int    `json:"age"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
	Password     string `json:"password"`
}

func Createnewuser(user *models.User) (*models.User, error) {
	// Check if the user provided all the required fields
	if user.Name == "" || user.Age == 0 || user.MobileNumber == "" || user.Email == "" || user.Password == "" {
		return nil, fmt.Errorf("all fields are required")
	}

	// Check if the email already exists in the database
	var ID int
	row := Db.QueryRow("SELECT id FROM users WHERE email=$1", user.Email)
	err := row.Scan(&ID)
	if err == nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Check if the mobile number is already registered in the database
	row = Db.QueryRow("SELECT id FROM users WHERE mobile_number=$1", user.MobileNumber)
	err = row.Scan(&ID)
	if err == nil {
		return nil, fmt.Errorf("mobile number already registered")
	}

	// Validate the age range
	if user.Age < 18 || user.Age > 100 {
		return nil, fmt.Errorf("age should be between 18 and 100")
	}

	// Hash the password and handle potential error
	hashedPassword := utils.HashPassword(user.Password)
	if hashedPassword == "" {
		return nil, fmt.Errorf("failed to hash password")
	}

	// Store the hashed password in the database
	sqlStatement := `INSERT INTO users (name, age, mobile_number, email, password) VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	err = Db.QueryRow(sqlStatement, user.Name, user.Age, user.MobileNumber, user.Email, hashedPassword).Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %v", err)
	}

	newUserResponse := &models.User{
		ID:           user.ID,
		Name:         user.Name,
		Age:          user.Age,
		MobileNumber: user.MobileNumber,
		Email:        user.Email,
	}

	// Return the newly created user and nil error
	return newUserResponse, nil
}

func UpdateUserByID(userID int, name string, age int) error {

	// SQL query to update the user's name and age by ID
	query := `
		UPDATE users
		SET name = $1, age = $2
		WHERE id = $3
		RETURNING id
	`

	// Execute the update query
	var updatedID int
	err := Db.QueryRow(query, name, age, userID).Scan(&updatedID)

	if err != nil {
		// Handle case where no rows are affected (user not found)
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		// Handle any other database errors
		return fmt.Errorf("could not update user: %v", err)
	}

	return nil
}
