package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/golang-jwt/jwt/v4"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

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

var db *sql.DB
var jwtKey = []byte("your_secret_key")

func init() {
	var err error
	cdb := "postgres://hemant:5689@localhost/postgres?sslmode=disable"
	db, err = sql.Open("postgres", cdb)

	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	// Confirming database connection
	fmt.Println("The database is connected")
}

func GenerateJWT(email string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Minute)
	claims := &Claims{
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
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	return claims, nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		claims, err := ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Add the claims to the context
		c.Set("email", claims.Email)
		c.Next()
	}
}

func getallusers(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	if page == "" {
		page = "1"
	}

	limit := c.DefaultQuery("limit", "10")
	if limit == "" {
		limit = "10"
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid page number"})
		return
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid limit number"})
		return
	}

	offset := (pageInt - 1) * limitInt

	var totalUsers int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not count users"})
		return
	}

	rows, err := db.Query(
		"SELECT id, name, age, mobile_number, email FROM users ORDER BY id LIMIT $1 OFFSET $2",
		limitInt, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Age, &user.MobileNumber, &user.Email); err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	c.Header("Page-Number", strconv.Itoa(pageInt))
	c.Header("User-Count", strconv.Itoa(len(users)))
	c.Header("Total-Users", strconv.Itoa(totalUsers))
	c.Header("Offset", strconv.Itoa(offset))
	c.Header("Limit", strconv.Itoa(limitInt))
	c.Header("Total-Pages", strconv.Itoa(totalUsers/limitInt))

	if users == nil || len(users) > totalUsers {
		c.JSON(http.StatusOK, map[string]any{"message": "No data available", "status": "failure"})
	} else {
		c.JSON(http.StatusOK, map[string]any{
			"message": "data is available",
			"status":  "success",
			"data":    users,
		})
	}
}

func getbyid(c *gin.Context) {
	id := c.Param("1")

	row := db.QueryRow(
		"SELECT * FROM users WHERE id=$1", id,
	)

	var user User
	if err := row.Scan(&user.ID, &user.Name, &user.Age, &user.MobileNumber, &user.Email, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, map[string]any{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		}
		return
	}

	if user.ID == 0 {
		c.JSON(http.StatusOK, map[string]any{"message": "No data available", "status": "failure"})
	} else {
		c.Header("Total-Count", "1")
		c.JSON(http.StatusOK, map[string]any{
			"message": "data is available",
			"status":  "success",
			"data":    user,
		})

	}
}

func newUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
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
	err = db.QueryRow(sqlStatement, user.Name, user.Age, user.MobileNumber, user.Email, hashedPassword).Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("New user created:", user)

	c.JSON(http.StatusCreated, user)

}

func login(c *gin.Context) {
	var user User
	var err error
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	row := db.QueryRow("SELECT * FROM users WHERE email=$1", user.Email)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, map[string]any{"error": "Invalid email or password"})
		return
	}

	var dbUser User
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
	tokenString, err := GenerateJWT(dbUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK,
		map[string]any{
			"status": "successfully logged in",
			"token":  tokenString})

	fmt.Println("User logged in:")
}

func main() {
	r := gin.Default()
	r.POST("/register", newUser)
	r.POST("/login", login)
	auth := r.Group("/")
	auth.Use(AuthMiddleware())
	{
		auth.GET("/users", getallusers)
		auth.GET("/users/:1", getbyid)
	}
	r.Run(":3000")
}
