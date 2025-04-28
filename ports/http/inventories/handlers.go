package inventories

import (
	"net/http"

	"github.com/devbenho/luka-platform/internal/inventory/dtos"
	"github.com/devbenho/luka-platform/internal/inventory/services"
	"github.com/devbenho/luka-platform/internal/utils"
	errors "github.com/devbenho/luka-platform/ports/http/errors"
	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	service services.IInventoryService
}

func NewInventoryHandler(service services.IInventoryService) *InventoryHandler {
	return &InventoryHandler{
		service: service,
	}
}

// @Summary Create a new inventory
// @Description Create a new inventory with the provided details
// @Tags inventories
// @Accept json
// @Produce json
// @Param inventory body dtos.CreateInventoryRequest true "Inventory details"
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /inventories [post]
func (h *InventoryHandler) Create(c *gin.Context) {
	var createInventoryRequest dtos.CreateInventoryRequest
	if err := c.ShouldBindJSON(&createInventoryRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}
	result, err := h.service.CreateInventory(c.Request.Context(), createInventoryRequest)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusCreated, "Inventory created successfully", result)
	c.JSON(http.StatusCreated, response)
}

// @Summary Update an inventory
// @Description Update an existing inventory
// @Tags inventories
// @Accept json
// @Produce json
// @Param id path string true "Inventory ID"
// @Param inventory body dtos.UpdateInventoryRequest true "Inventory update data"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /inventories/{id} [patch]
func (h *InventoryHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var updateInventoryRequest dtos.UpdateInventoryRequest
	if err := c.ShouldBindJSON(&updateInventoryRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}
	inventory, err := h.service.UpdateInventory(c.Request.Context(), id, updateInventoryRequest)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Inventory updated successfully", inventory)
	c.JSON(http.StatusOK, response)
}

// @Summary Get inventory by ID
// @Description Get detailed information about a specific inventory
// @Tags inventories
// @Produce json
// @Param id path string true "Inventory ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /inventories/{id} [get]
func (h *InventoryHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	inventory, err := h.service.GetInventoryByID(c.Request.Context(), id)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Inventory fetched successfully", inventory)
	c.JSON(http.StatusOK, response)
}

// @Summary Delete an inventory
// @Description Delete an inventory from the system
// @Tags inventories
// @Produce json
// @Param id path string true "Inventory ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /inventories/{id} [delete]
func (h *InventoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.service.DeleteInventory(c.Request.Context(), id)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Inventory deleted successfully", nil)
	c.JSON(http.StatusOK, response)
}
