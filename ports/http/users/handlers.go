package users

import (
	dtos "github.com/devbenho/luka-platform/internal/user/dtos/users"
	"github.com/devbenho/luka-platform/internal/user/services"
	apperror "github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/response"
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
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /user [post]
func (h *UserHandler) Register(c *gin.Context) error {
	var createUserRequest dtos.CreateUserRequest

	if err := c.ShouldBindJSON(&createUserRequest); err != nil {
		return response.ErrorBuilder(apperror.BadRequest(err)).Send(c)
	}

	// validate the request body
	if err := createUserRequest.Validate(); err != nil {
		return response.ErrorBuilder(apperror.BadRequest(err)).Send(c)
	}

	result, err := h.service.Register(c.Request.Context(), &createUserRequest)
	if err != nil {
		return response.ErrorBuilder(err).Send(c)
	}

	return response.SuccessBuilder(result).Send(c)
}

// Login handles user login requests
// @Summary Login a user
// @Description Login a user with the provided credentials
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body dtos.AuthDTO true "User credentials"
// @Success 200 {object} dtos.AuthResponseDTO
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /auth/login [post]
// func (h *UserHandler) Login(c *gin.Context) {
// 	var authDTO dtos.LoginRequest

// 	if err := c.ShouldBindJSON(&authDTO); err != nil {
// 		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
// 		return
// 	}

// 	response, err := h.service.Login(c.Request.Context(), &authDTO)
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse(http.StatusUnauthorized, "Failed to login", err.Error()))
// 		return
// 	}

// 	result := utils.NewSuccessResponse(http.StatusOK, "User logged in successfully", response)
// 	c.JSON(http.StatusOK, result)
// }

// GetUserByID handles requests to fetch user details by ID
// @Summary Get user by ID
// @Description Get user details by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dtos.UserResponseDTO
// @Failure 404 {object} utils.ErrorResponse
// @Router /user/{id} [get]
// func (h *UserHandler) GetUserByID(c *gin.Context) {
// 	id := c.Param("id") // Get the user ID from the request path

// 	user, err := h.service.GetUserByID(c.Request.Context(), id)
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, utils.NewErrorResponse(http.StatusNotFound, "User not found", err.Error()))
// 		return
// 	}

// 	c.JSON(http.StatusOK, user)
// }

// UpdateUser handles user update requests
// @Summary Update user details
// @Description Update user details by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body dtos.UpdateUserRequest true "Updated user details"
// @Success 200 {object} dtos.UserResponseDTO
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /user/{id} [put]
// func (h *UserHandler) UpdateUser(c *gin.Context) {
// 	id := c.Param("id")

// 	var updateUserRequest dtos.UpdateUserRequest
// 	if err := c.ShouldBindJSON(&updateUserRequest); err != nil {
// 		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
// 		return
// 	}

// 	updatedUser, err := h.service.UpdateUser(c.Request.Context(), id, &updateUserRequest)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Failed to update user", err.Error()))
// 		return
// 	}

// 	c.JSON(http.StatusOK, updatedUser)
// }

// DeleteUser handles user deletion requests
// @Summary Delete user
// @Description Delete user by ID
// @Tags users
// @Param id path string true "User ID"
// @Success 204
// @Failure 500 {object} utils.ErrorResponse
// @Router /user/{id} [delete]
// func (h *UserHandler) DeleteUser(c *gin.Context) {
// 	id := c.Param("id")

// 	if err := h.service.DeleteUser(c.Request.Context(), id); err != nil {
// 		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Failed to delete user", err.Error()))
// 		return
// 	}

// 	c.Status(http.StatusNoContent)
// }
