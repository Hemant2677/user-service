package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var Db *sql.DB

func init() {
	var err error
	cdb := "postgres://hemant:5689@localhost/postgres?sslmode=disable"
	Db, err = sql.Open("postgres", cdb)

	if err != nil {
		panic(err)
	}

	if err = Db.Ping(); err != nil {
		panic(err)
	}
	// Confirming database connection
	fmt.Println("The database is connected")
}

// CreateUser creates a new user in the database
func Getallusers(page int, limit int) ([]UserResponse, int, error) {
	if page < 1 {
		return nil, 0, fmt.Errorf("invalid page number")
	}

	if limit < 1 {
		return nil, 0, fmt.Errorf("invalid limit number")
	}

	offset := (page - 1) * limit

	// Get the total number of users
	var totalUsers int
	err := Db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return nil, 0, fmt.Errorf("could not count users: %v", err)
	}

	// Fetch the user data with pagination
	rows, err := Db.Query(
		"SELECT id, name, age, mobile_number, email FROM users ORDER BY id LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Prepare a list to store users
	var users []UserResponse
	for rows.Next() {
		var user UserResponse
		if err := rows.Scan(&user.ID, &user.Name, &user.Age, &user.MobileNumber, &user.Email); err != nil {
			return nil, 0, err
		}

		users = append(users, user)
	}

	// Check for any row scanning errors
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	// Return the list of users, total users count, and nil error if successful
	return users, totalUsers, nil
}

func Getuserbyid(id string) (*UserResponse, error) {
	var user UserResponse
	row := Db.QueryRow("SELECT id, name, age, mobile_number, email FROM users WHERE id=$1", id)

	err := row.Scan(&user.ID, &user.Name, &user.Age, &user.MobileNumber, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return &UserResponse{}, fmt.Errorf("user not found")
		}
		return &UserResponse{}, err
	}

	return &user, nil
}
