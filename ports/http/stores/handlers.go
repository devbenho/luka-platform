package stores

import (
	"net/http"

	"github.com/devbenho/luka-platform/internal/store/dtos"
	"github.com/devbenho/luka-platform/internal/store/services"
	"github.com/devbenho/luka-platform/internal/utils"
	"github.com/devbenho/luka-platform/pkg/slug"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StoreHandler struct {
	service services.IStoreService
}

func NewStoreHandler(service services.IStoreService) *StoreHandler {
	return &StoreHandler{
		service: service,
	}
}

// Create handles user registration requests
// @Summary Create a new user
// @Description Create a new user with the provided details
// @Tags users
// @Accept json
// @Produce json
// @Param user body dtos.CreateUserRequest true "User details"
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /user [post]
func (h *StoreHandler) Create(c *gin.Context) {
	value, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse(http.StatusUnauthorized, "Unauthorized", "Unauthorized"))
		return
	}

	userIdStr, ok := value.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Invalid user ID", "Invalid user ID"))
		return
	}

	userId, err := primitive.ObjectIDFromHex(userIdStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Invalid user ID format", err.Error()))
		return
	}

	var createStoreRequest dtos.CreateStoreRequest

	if err := c.ShouldBindJSON(&createStoreRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}
	createStoreRequest.OwnerId = userId
	createStoreRequest.Slug = slug.GenerateSlug(createStoreRequest.Name)

	result, err := h.service.CreateStore(c.Request.Context(), &createStoreRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Failed to register user", err.Error()))
		return
	}

	response := utils.NewSuccessResponse(http.StatusCreated, "Store registered successfully", result)
	c.JSON(http.StatusCreated, response)
}
