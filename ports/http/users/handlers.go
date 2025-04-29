package users

import (
	"fmt"
	"log"
	"net/http"

	dtos "github.com/devbenho/luka-platform/internal/user/dtos/users"
	"github.com/devbenho/luka-platform/internal/user/services"
	"github.com/devbenho/luka-platform/internal/utils"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service services.IUserService
}

func NewUserHandler(service services.IUserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// Register handles user registration requests
// @Summary Register a new user
// @Description Register a new user with the provided details
// @Tags users
// @Accept json
// @Produce json
// @Param user body dtos.CreateUserRequest true "User details"
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	fmt.Println("register user handler")
	var createUserRequest dtos.CreateUserRequest

	if err := c.ShouldBindJSON(&createUserRequest); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}

	log.Println(createUserRequest)

	result, err := h.service.Register(c.Request.Context(), &createUserRequest)
	log.Println(result)
	log.Println(err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Failed to register user", err.Error()))
		return
	}

	response := utils.NewSuccessResponse(http.StatusCreated, "User registered successfully", result)
	c.JSON(http.StatusCreated, response)
}

// Login handles user login requests
// @Summary Login a user
// @Description Login a user with the provided credentials
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body dtos.AuthDTO true "User credentials"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var authDTO dtos.AuthDTO

	if err := c.ShouldBindJSON(&authDTO); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}

	response, err := h.service.Login(c.Request.Context(), &authDTO)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse(http.StatusUnauthorized, "Failed to login", err.Error()))
		return
	}

	result := utils.NewSuccessResponse(http.StatusOK, "User logged in successfully", response)
	c.JSON(http.StatusOK, result)
}

// GetUserByID handles requests to fetch user details by ID
// @Summary Get user by ID
// @Description Get user details by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse(http.StatusNotFound, "User not found", err.Error()))
		return
	}
	log.Printf("user: %+v", user)
	c.JSON(http.StatusOK, user)
}

// UpdateUser handles user update requests
// @Summary Update user details
// @Description Update user details by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body dtos.UpdateUserRequest true "Updated user details"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var updateUserRequest dtos.UpdateUserRequest
	if err := c.ShouldBindJSON(&updateUserRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}

	updatedUser, err := h.service.UpdateUser(c.Request.Context(), id, &updateUserRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Failed to update user", err.Error()))
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// DeleteUser handles user deletion requests
// @Summary Delete user
// @Description Delete user by ID
// @Tags users
// @Param id path string true "User ID"
// @Success 204
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Failed to delete user", err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
