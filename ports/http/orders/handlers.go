package orders

import (
	"net/http"

	"github.com/devbenho/luka-platform/internal/orders/models"
	dtos "github.com/devbenho/luka-platform/internal/orders/order_dtos"
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

// @Summary Create a new order
// @Description Create a new order with the provided details
// @Tags orders
// @Accept json
// @Produce json
// @Param order body dtos.CreateOrderRequest true "Order details"
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
	var createOrderRequest dtos.CreateOrderRequest
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

// @Summary Update order status
// @Description Update the status of an existing order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param status body models.OrderStatus true "New order status"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /orders/{id}/status [patch]
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

// @Summary Get order by ID
// @Description Get detailed information about a specific order
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /orders/{id} [get]
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

// @Summary List orders
// @Description Get a list of orders, optionally filtered by customer ID
// @Tags orders
// @Produce json
// @Param customer_id query string false "Filter orders by customer ID"
// @Success 200 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /orders [get]
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
