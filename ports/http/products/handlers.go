package products

import (
	"net/http"

	"github.com/devbenho/luka-platform/internal/product/dtos"
	"github.com/devbenho/luka-platform/internal/product/services"
	"github.com/devbenho/luka-platform/internal/utils"
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

// @Summary Create a new product
// @Description Create a new product in the system
// @Tags products
// @Accept json
// @Produce json
// @Param product body dtos.CreateProductRequest true "Product Data"
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req dtos.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid request body", err.Error()))
		return
	}

	product, err := h.service.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Failed to create product", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.NewSuccessResponse(http.StatusCreated, "Product created successfully", product))
}

// @Summary Update a product
// @Description Update an existing product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body dtos.UpdateProductRequest true "Product Update Data"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /products/{id} [patch]
func (h *ProductHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dtos.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid request body", err.Error()))
		return
	}

	product, err := h.service.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Failed to update product", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(http.StatusOK, "Product updated successfully", product))
}

// @Summary Get a product by ID
// @Description Get detailed information about a specific product
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /products/{id} [get]
func (h *ProductHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	product, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Failed to get product", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(http.StatusOK, "Product retrieved successfully", product))
}

// @Summary Delete a product
// @Description Delete a product from the system
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.service.DeleteProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, "Failed to delete product", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(http.StatusOK, "Product deleted successfully", nil))
}
