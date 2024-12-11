package inventories

import (
	"log"
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

func (h *InventoryHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var updateInventoryRequest dtos.UpdateInventoryRequest
	if err := c.ShouldBindJSON(&updateInventoryRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}
	log.Println(`The update inventory request is `, updateInventoryRequest)
	inventory, err := h.service.UpdateInventory(c.Request.Context(), id, updateInventoryRequest)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Inventory updated successfully", inventory)
	c.JSON(http.StatusOK, response)
}

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
