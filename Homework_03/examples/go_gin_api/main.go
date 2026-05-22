package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	r := gin.Default()

	r.GET("/users", listUsers)
	r.GET("/users/:id", getUser)
	r.POST("/users", createUser)
	r.DELETE("/users/:id", deleteUser)

	_ = r.Run(":8080")
}

// listUsers returns all users.
func listUsers(c *gin.Context) {
	c.JSON(http.StatusOK, []User{})
}

// getUser returns a single user by id.
func getUser(c *gin.Context) {
	c.JSON(http.StatusOK, User{ID: 1})
}

// createUser creates a new user.
func createUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, User{ID: 1, Name: req.Name, Email: req.Email})
}

// deleteUser removes a user by id.
func deleteUser(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
