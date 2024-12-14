package orders

import (
	"net/http"

	"github.com/devbenho/luka-platform/internal/orders/models"
	orders "github.com/devbenho/luka-platform/internal/orders/order_dtos"
	"github.com/devbenho/luka-platform/internal/orders/services"
	"github.com/devbenho/luka-platform/internal/utils"
	errors "github.com/devbenho/luka-platform/ports/http/errors"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service services.IOrderService
}

func NewOrderHandler(service services.IOrderService) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var createOrderRequest orders.CreateOrderRequest
	if err := c.ShouldBindJSON(&createOrderRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid input", err.Error()))
		return
	}

	result, err := h.service.CreateOrder(c.Request.Context(), createOrderRequest)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusCreated, "Order created successfully", result)
	c.JSON(http.StatusCreated, response)
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var status models.OrderStatus
	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "Invalid status", err.Error()))
		return
	}

	err := h.service.UpdateOrderStatus(c.Request.Context(), id, status)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Order status updated successfully", nil)
	c.JSON(http.StatusOK, response)
}

func (h *OrderHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Order fetched successfully", order)
	c.JSON(http.StatusOK, response)
}

func (h *OrderHandler) List(c *gin.Context) {
	customerID := c.Query("customer_id")
	orders, err := h.service.ListOrders(c.Request.Context(), customerID)
	if err != nil {
		apiError := errors.MapErrorToHTTP(err)
		c.JSON(apiError.Status, apiError)
		return
	}

	response := utils.NewSuccessResponse(http.StatusOK, "Orders fetched successfully", orders)
	c.JSON(http.StatusOK, response)
}
