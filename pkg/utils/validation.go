package utils

import (
	"errors"

	"github.com/golang-jwt/jwt"
)

//creting a function that extract id, name and email from token

func ExtractUserInfo(tokenString string) (int, string, string, error) {
	claims := &Claims{}

	// Parse the token with the custom claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	// Handle invalid token or parsing error
	if err != nil || !token.Valid {
		return 0, "", "", errors.New("invalid token")
	}

	// Return extracted ID, Name, and Email
	return claims.ID, claims.Name, claims.Email, nil
}


