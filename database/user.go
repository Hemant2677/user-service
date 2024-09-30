package database

import (
	"database/sql"
	"net/http"
	"strconv"

	"user-service/models"

	"github.com/gin-gonic/gin"
)

func Getallusers(c *gin.Context) {
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
	err = Db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not count users"})
		return
	}

	rows, err := Db.Query(
		"SELECT id, name, age, mobile_number, email FROM users ORDER BY id LIMIT $1 OFFSET $2",
		limitInt, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User
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

func Byid(c *gin.Context) {
	id := c.Param("id")

	row := Db.QueryRow(
		"SELECT * FROM users WHERE id=$1", id,
	)

	var user models.User
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
