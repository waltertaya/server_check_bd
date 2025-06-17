package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/waltertaya/server_check_bd/internal/logger"
	"github.com/waltertaya/server_check_bd/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandlers handles authentication-related HTTP requests
type AuthHandlers struct {
	db *sqlx.DB
}

// NewAuthHandlers creates a new auth handlers instance
func NewAuthHandlers(db *sqlx.DB) *AuthHandlers {
	return &AuthHandlers{
		db: db,
	}
}

// RegisterRequest represents the request to register a new user
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents the request to login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterHandler handles user registration
func (h *AuthHandlers) RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid registration request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check if email already exists
	var existingUser models.User
	err := h.db.Get(&existingUser, "SELECT id FROM users WHERE email = ?", req.Email)
	if err == nil {
		logger.Error("Email %s already exists", req.Email)
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Create user
	user := models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	_, err = h.db.NamedExec(`
		INSERT INTO users (email, password)
		VALUES (:email, :password)
	`, user)
	if err != nil {
		logger.Error("Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	logger.Info("User with email %s registered successfully", req.Email)
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// LoginHandler handles user login
func (h *AuthHandlers) LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid login request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get user
	var user models.User
	err := h.db.Get(&user, "SELECT * FROM users WHERE email = ?", req.Email)
	if err != nil {
		logger.Error("User with email %s not found: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		logger.Error("Invalid password for user with email %s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	logger.Info("User with email %s logged in successfully", req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}
