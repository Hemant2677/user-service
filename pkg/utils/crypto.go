package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("your_secret_key")

//create a new function that take return hashpassword using bcrypt algorithm

func HashPassword(password string) string {
	hash := md5.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

func ComparePasswords(hashedPassword, password string) error {
	// Create a new MD5 hash object
	hash := md5.New()
	// Write the password to the hash object
	hash.Write([]byte(password))
	// Get the hexadecimal representation of the hash
	hashedInput := hex.EncodeToString(hash.Sum(nil))
	// Compare the hashed input with the stored hashed password
	if hashedInput != hashedPassword {
		return fmt.Errorf("invalid password")
	}
	return nil
}

type Claims struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token for the user
func GenerateJWT(ID int, Name string, Email string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := Claims{
		ID:    ID,
		Name:  Name,
		Email: Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates the JWT token
func ValidateJWT(tokenString string) (Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return Claims{}, err
	}

	if !token.Valid {
		return Claims{}, err
	}

	return claims, nil
}

