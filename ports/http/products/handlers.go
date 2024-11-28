package products

import (
	"net/http"

	"github.com/devbenho/luka-platform/internal/product/dtos"
	"github.com/devbenho/luka-platform/internal/product/services"
	"github.com/devbenho/luka-platform/internal/utils"
	errors "github.com/devbenho/luka-platform/ports/http/errors"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	service services.IProductService
}

func NewProductHandler(service services.IProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

// Create handles product creation requests
// @Summary Create a new product
// @Description Create a new product with the provided details
// @Tags products
// @Accept json
// @Produce json
// @Param product body dtos.CreateProductRequest true "Product details"
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var createProductRequest dtos.CreateProductRequest

	if err := c.ShouldBindJSON(&createProductRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}

	result, err := h.service.CreateProduct(c.Request.Context(), &createProductRequest)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusCreated, "Product created successfully", result)
	c.JSON(http.StatusCreated, response)
}

// GetById handles fetching a product by ID
// @Summary Get a product by ID
// @Description Fetch a product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /products/{id} [get]
func (h *ProductHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	product, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Product fetched successfully", product)
	c.JSON(http.StatusOK, response)
}

// Update handles updating a product
// @Summary Update a product
// @Description Update a product with the provided details
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body dtos.UpdateProductRequest true "Product details"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /products/{id} [patch]
func (h *ProductHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var updateProductRequest dtos.UpdateProductRequest
	if err := c.ShouldBindJSON(&updateProductRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}

	product, err := h.service.UpdateProduct(c.Request.Context(), id, &updateProductRequest)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Product updated successfully", product)
	c.JSON(http.StatusOK, response)
}

// Delete handles deleting a product
// @Summary Delete a product
// @Description Delete a product with the provided ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.service.DeleteProduct(c.Request.Context(), id)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Product deleted successfully", nil)
	c.JSON(http.StatusOK, response)
}
