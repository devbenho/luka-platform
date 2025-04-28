package stores

import (
	"net/http"

	"github.com/devbenho/luka-platform/internal/store/dtos"
	"github.com/devbenho/luka-platform/internal/store/services"
	"github.com/devbenho/luka-platform/internal/utils"
	"github.com/devbenho/luka-platform/pkg/slug"
	errors "github.com/devbenho/luka-platform/ports/http/errors"
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

// Create handles store creation requests
// @Summary Create a new store
// @Description Create a new store with the provided details
// @Tags stores
// @Accept json
// @Produce json
// @Param store body dtos.CreateStoreRequest true "Store details"
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /stores [post]
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

// GetById handles fetching a store by ID
// @Summary Get a store by ID
// @Description Fetch a store by ID
// @Tags stores
// @Produce json
// @Param id path string true "Store ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /stores/{id} [get]
func (h *StoreHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	store, err := h.service.GetStoreByID(c.Request.Context(), id)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Store fetched successfully", store)
	c.JSON(http.StatusOK, response)
}

// Update handles updating a store
// @Summary Update a store
// @Description Update a store with the provided details
// @Tags stores
// @Accept json
// @Produce json
// @Param id path string true "Store ID"
// @Param store body dtos.UpdateStoreRequest true "Store details"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /stores/{id} [put]
func (h *StoreHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var updateStoreRequest dtos.UpdateStoreRequest
	if err := c.ShouldBindJSON(&updateStoreRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}

	store, err := h.service.UpdateStore(c.Request.Context(), id, &updateStoreRequest)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Store updated successfully", store)
	c.JSON(http.StatusOK, response)
}

// Delete handles deleting a store
// @Summary Delete a store
// @Description Delete a store with the provided ID
// @Tags stores
// @Produce json
// @Param id path string true "Store ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /stores/{id} [delete]
func (h *StoreHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.service.DeleteStore(c.Request.Context(), id)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Store deleted successfully", nil)
	c.JSON(http.StatusOK, response)
}
