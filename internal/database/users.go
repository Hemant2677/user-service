package database

import (
	"database/sql"
	"fmt"
	"log"
	"user-service/pkg/utils"
)

type User struct {
	ID           int
	Name         string `json:"name"`
	Age          int    `json:"age"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
	Password     string `json:"password"`
}

type UserResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Age          int    `json:"age"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
}

type userRequest struct {
	Name         string `json:"name"`
	Age          int    `json:"age"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
	Password     string `json:"password"`
}

func Createnewuser(user *User) (*User, error) {

	// Check if the user provided all the required fields

	if user.Name == "" || user.Age == 0 || user.MobileNumber == "" || user.Email == "" || user.Password == "" {
		return &User{}, fmt.Errorf("all fields are required")
	}

	// Check if the email already exists in the database
	var ID int
	row := Db.QueryRow("SELECT id FROM users WHERE email=$1", user.Email)
	err := row.Scan(&ID)
	if err == nil {
		return &User{}, fmt.Errorf("email already exists")
	}

	// Check if the mobile number is already registered in the database
	row = Db.QueryRow("SELECT id FROM users WHERE mobile_number=$1", user.MobileNumber)
	err = row.Scan(&ID)
	if err == nil {
		return &User{}, fmt.Errorf("mobile number already registered")
	}

	// Validate the age range
	if user.Age < 18 || user.Age > 100 {
		return &User{}, fmt.Errorf("age should be between 18 and 100")
	}

	// Hash the password and handle potential error
	hashedPassword := utils.HashPassword(user.Password)
	if hashedPassword == "" {
		return &User{}, fmt.Errorf("failed to hash password")
	}

	// Store the hashed password in the database
	sqlStatement := `INSERT INTO users (name, age, mobile_number, email, password) VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	err = Db.QueryRow(sqlStatement, user.Name, user.Age, user.MobileNumber, user.Email, hashedPassword).Scan(&user.ID)
	if err != nil {
		return &User{}, fmt.Errorf("failed to insert user: %v", err)
	}

	newUserResponse := &User{
		ID:           user.ID,
		Name:         user.Name,
		Age:          user.Age,
		MobileNumber: user.MobileNumber,
		Email:        user.Email,
	}

	// Return the newly created user and nil error
	return newUserResponse, nil
}

func Login(user userRequest) (string, error) {
	log.Printf("User login attempt: email=%s, password=%s", user.Email, user.Password)
	// Check if the user provided both email and password
	if user.Email == "" || user.Password == "" {
		return "", fmt.Errorf("email and password are required")
	}

	// Query the database to find the user by email
	row := Db.QueryRow("SELECT * FROM users WHERE email=$1", user.Email)

	var dbPassword string
	err := row.Scan(&dbPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return an error if no user with the given email exists
			return "", fmt.Errorf("invalid email or password")
		}
		// Return a general error if the database query fails
		return "", err
	}

	// Compare the provided password with the hashed password in the database
	if err := utils.ComparePasswords(dbPassword, user.Password); err != nil {
		// Return an error if the passwords do not match
		return "", fmt.Errorf("invalid password")
	}

	// If the passwords match, generate a JWT token for the authenticated user
	tokenString, err := utils.GenerateJWT(user.Email)
	if err != nil {
		// Return an error if token generation fails
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

	// Return the generated token and no error if everything is successful
	return tokenString, nil
}
