package middleware

import (
	"net/http"
	"user-service/pkg/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, map[string]any{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, map[string]any{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Add the claims to the context
		c.Set("email", claims.Email)
		c.Next()
	}
}
