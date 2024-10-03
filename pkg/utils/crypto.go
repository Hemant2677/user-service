package utils

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("your_secret_key")

//create a new function that take return hashpassword using bcrypt algorithm

func HashPassword(password string) string {
	hash := md5.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

func ComparePasswords(hashedPassword, password string) error {
	// Use bcrypt to compare the hashed password with the plain password
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token for the user
func GenerateJWT(email string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := Claims{
		Email: email,
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
