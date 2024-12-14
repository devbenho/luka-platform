package orders

import (
	configs "github.com/devbenho/luka-platform/configs"
	"github.com/devbenho/luka-platform/internal/inventory/repositories"
	"github.com/devbenho/luka-platform/internal/inventory/services"
	orderRepo "github.com/devbenho/luka-platform/internal/orders/repositories"
	orderSvc "github.com/devbenho/luka-platform/internal/orders/services"
	productRepo "github.com/devbenho/luka-platform/internal/product/repositories"
	productSvc "github.com/devbenho/luka-platform/internal/product/services"
	storeRepo "github.com/devbenho/luka-platform/internal/store/repositories"
	"github.com/devbenho/luka-platform/pkg/database"
	middleware "github.com/devbenho/luka-platform/pkg/middlewares"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.RouterGroup, mongoDb database.IDatabase, validator *validation.Validator, config configs.Config) {
	// Initialize repositories
	orderRepository := orderRepo.NewOrderRepository(mongoDb)
	inventoryRepository := repositories.NewInventoryRepository(mongoDb)
	productRepository := productRepo.NewProductRepository(mongoDb)
	storeRepository := storeRepo.NewStoreRepository(mongoDb)
	// Initialize services
	inventoryService := services.NewInventoryService(inventoryRepository, validator)
	productService := productSvc.NewProductService(productRepository, storeRepository, validator)
	orderService := orderSvc.NewOrderService(orderRepository, inventoryService, productService, validator)

	// Initialize handler
	orderHandler := NewOrderHandler(orderService)

	// Define routes
	ordersRoute := r.Group("/orders")
	{
		ordersRoute.POST("/", middleware.JWTAuth(), orderHandler.Create)
		ordersRoute.GET("/:id", middleware.JWTAuth(), orderHandler.GetById)
		ordersRoute.GET("/", middleware.JWTAuth(), orderHandler.List)
		ordersRoute.PATCH("/:id/status", middleware.JWTAuth(), orderHandler.UpdateStatus)
	}
}
