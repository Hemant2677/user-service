package models

type User struct {
	ID           int
	Name         string `json:"name"`
	Age          int    `json:"age"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
	Password     string `json:"password"`
}