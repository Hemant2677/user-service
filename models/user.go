package models

import "github.com/golang-jwt/jwt/v4"

type User struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Age          int    `json:"age"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
	Password     string `json:"-"`
}

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}
