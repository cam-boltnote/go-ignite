package routes

import (
	"strconv"

	"github.com/cam-boltnote/go-ignite/internal/middleware"
	"github.com/cam-boltnote/go-ignite/internal/models"
	"github.com/cam-boltnote/go-ignite/internal/services"

	"github.com/gin-gonic/gin"
)

// UserRoutes handles all user-related routes
type UserRoutes struct {
	userService *services.UserService
}

// NewUserRoutes creates a new user routes instance
func NewUserRoutes(userService *services.UserService) *UserRoutes {
	return &UserRoutes{
		userService: userService,
	}
}

// RegisterPublicRoutes registers public user-related routes
func (r *UserRoutes) RegisterPublicRoutes(rg *gin.RouterGroup) {
	// Public routes (no authentication required)
	rg.OPTIONS("/user", middleware.CorsOptionsHandler)
	rg.POST("/user", r.CreateUser)

	rg.OPTIONS("/user/login", middleware.CorsOptionsHandler)
	rg.POST("/user/login", r.Login)
}

// RegisterRoutes registers protected user-related routes
func (r *UserRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/user")
	{
		users.OPTIONS("/:id", middleware.CorsOptionsHandler)
		users.GET("/:id", r.GetUser)
		users.PUT("/:id", r.UpdateUser)
		users.DELETE("/:id", r.DeleteUser)

		users.OPTIONS("/email/:email", middleware.CorsOptionsHandler)
		users.GET("/email/:email", r.GetUserByEmail)

		users.OPTIONS("/:id/password", middleware.CorsOptionsHandler)
		users.PUT("/:id/password", r.UpdatePassword)

		users.OPTIONS("/:id/activate", middleware.CorsOptionsHandler)
		users.PUT("/:id/activate", r.ActivateUser)

		users.OPTIONS("/:id/deactivate", middleware.CorsOptionsHandler)
		users.PUT("/:id/deactivate", r.DeactivateUser)
	}
}

// CreateUser handles user registration
func (r *UserRoutes) CreateUser(c *gin.Context) {
	var input services.CreateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user, err := r.userService.CreateUser(input)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, user)
}

// Login handles user authentication
func (r *UserRoutes) Login(c *gin.Context) {
	var loginInput struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginInput); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user, err := r.userService.ValidateCredentials(loginInput.Email, loginInput.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error generating token"})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
		"user":  user,
	})
}

// GetUser retrieves a user by ID
func (r *UserRoutes) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := r.userService.GetByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, user)
}

// UpdateUser updates a user's information
func (r *UserRoutes) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user.ID = uint(id)
	if err := r.userService.Update(&user); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, user)
}

// GetUserByEmail retrieves a user by email
func (r *UserRoutes) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")

	user, err := r.userService.GetByEmail(email)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, user)
}

// DeleteUser deletes a user
func (r *UserRoutes) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := r.userService.Delete(uint(id)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User deleted successfully"})
}

// UpdatePassword updates a user's password
func (r *UserRoutes) UpdatePassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	var input struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Validate current password
	user, err := r.userService.GetByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	// Note: In a real application, you would:
	// 1. Hash the current password
	// 2. Compare it with the stored hash
	// 3. If they match, hash the new password
	// 4. Update the stored hash
	if user.Password != input.CurrentPassword {
		c.JSON(401, gin.H{"error": "Current password is incorrect"})
		return
	}

	if err := r.userService.UpdatePassword(uint(id), input.NewPassword); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Password updated successfully"})
}

// ActivateUser activates a user account
func (r *UserRoutes) ActivateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := r.userService.Activate(uint(id)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User account activated successfully"})
}

// DeactivateUser deactivates a user account
func (r *UserRoutes) DeactivateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := r.userService.Deactivate(uint(id)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User account deactivated successfully"})
}
