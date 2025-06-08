package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waltertaya/server_check_bd/models"
	"github.com/waltertaya/server_check_bd/db" // sqlite database(connected using sqlx)
)

func LoginHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validate user credentials
	var dbUser models.User
	err := db.DB.Get(&dbUser, "SELECT id, username, email FROM users WHERE username = ? AND password = ?", user.Username, user.Password)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	// If credentials are valid, return user info (excluding password)
	if dbUser.ID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	user.ID = dbUser.ID
	user.Username = dbUser.Username
	user.Email = dbUser.Email
	user.Password = ""
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user": user})
}

func RegisterHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Check if username already exists
	var existingUser models.User
	err := db.DB.Get(&existingUser, "SELECT id FROM users WHERE username = ?", user.Username)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Insert new user into the database
	_, err = db.DB.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", user.Username, user.Email, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}
