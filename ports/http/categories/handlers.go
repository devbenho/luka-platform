package categories

import (
	"net/http"

	"github.com/devbenho/luka-platform/internal/category/dtos"
	"github.com/devbenho/luka-platform/internal/category/services"
	"github.com/devbenho/luka-platform/internal/utils"
	"github.com/devbenho/luka-platform/pkg/slug"
	errors "github.com/devbenho/luka-platform/ports/http/errors"
	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	service services.ICategoryService
}

func NewCategoryHandler(service services.ICategoryService) *CategoryHandler {
	return &CategoryHandler{
		service: service,
	}
}

// Create handles category creation requests
// @Summary Create a new category
// @Description Create a new category with the provided details
// @Tags categories
// @Accept json
// @Produce json
// @Param category body dtos.CreateCategoryRequest true "Category details"
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var createCategoryRequest dtos.CreateCategoryRequest

	if err := c.ShouldBindJSON(&createCategoryRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}
	createCategoryRequest.Slug = slug.GenerateSlug(createCategoryRequest.Name)
	result, err := h.service.CreateCategory(c.Request.Context(), &createCategoryRequest)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusCreated, "Category created successfully", result)
	c.JSON(http.StatusCreated, response)
}

// GetById handles fetching a category by ID
// @Summary Get a category by ID
// @Description Fetch a category by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /categories/{id} [get]
func (h *CategoryHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	category, err := h.service.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}
	category.Slug = slug.GenerateSlug(category.Name)
	response := utils.NewSuccessResponse(http.StatusOK, "Category fetched successfully", category)
	c.JSON(http.StatusOK, response)
}

// Update handles updating a category
// @Summary Update a category
// @Description Update a category with the provided details
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param category body dtos.UpdateCategoryRequest true "Category details"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /categories/{id} [patch]
func (h *CategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var updateCategoryRequest dtos.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&updateCategoryRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}
	slugValue := slug.GenerateSlug(*updateCategoryRequest.Name)
	updateCategoryRequest.Slug = &slugValue
	category, err := h.service.UpdateCategory(c.Request.Context(), id, &updateCategoryRequest)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Category updated successfully", category)
	c.JSON(http.StatusOK, response)
}

// Delete handles deleting a category
// @Summary Delete a category
// @Description Delete a category with the provided ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.service.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Category deleted successfully", nil)
	c.JSON(http.StatusOK, response)
}
