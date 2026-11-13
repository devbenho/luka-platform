package services

import (
	"context"
	stderrors "errors"
	"testing"

	inventoryDTOs "github.com/devbenho/luka-platform/internal/inventory/dtos"
	inventoryModels "github.com/devbenho/luka-platform/internal/inventory/models"
	"github.com/devbenho/luka-platform/internal/orders/models"
	dtos "github.com/devbenho/luka-platform/internal/orders/order_dtos"
	productDTOs "github.com/devbenho/luka-platform/internal/product/dtos"
	productModels "github.com/devbenho/luka-platform/internal/product/models"
	apperrors "github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- fakes ---------------------------------------------------------------

type fakeOrderRepo struct {
	createErr error
	created   *models.Order
	orders    map[string]*models.Order
	updated   map[string]models.OrderStatus
}

func (r *fakeOrderRepo) CreateOrder(_ context.Context, order *models.Order) (*models.Order, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	order.ID = primitive.NewObjectID()
	r.created = order
	return order, nil
}

func (r *fakeOrderRepo) UpdateOrder(_ context.Context, id string, order *models.Order) error {
	if r.updated == nil {
		r.updated = map[string]models.OrderStatus{}
	}
	r.updated[id] = order.Status
	return nil
}

func (r *fakeOrderRepo) GetOrderByID(_ context.Context, id string) (*models.Order, error) {
	return r.orders[id], nil
}

func (r *fakeOrderRepo) ListOrders(context.Context, string) ([]models.Order, error) {
	return nil, nil
}

type inventoryUpdate struct {
	id       string
	quantity int
}

// fakeInventory stores stock keyed by the ID string it is called with and
// records every UpdateInventory call so tests can assert which ID was used.
type fakeInventory struct {
	stock       map[string]int
	updates     []inventoryUpdate
	rollbackErr error // returned for positive (restoring) quantity updates
}

func (f *fakeInventory) CreateInventory(context.Context, inventoryDTOs.CreateInventoryRequest) (*inventoryModels.Inventory, error) {
	return nil, nil
}

func (f *fakeInventory) UpdateInventory(_ context.Context, id string, dto inventoryDTOs.UpdateInventoryRequest) (*inventoryModels.Inventory, error) {
	f.updates = append(f.updates, inventoryUpdate{id: id, quantity: *dto.Quantity})
	if *dto.Quantity > 0 && f.rollbackErr != nil {
		return nil, f.rollbackErr
	}
	if _, ok := f.stock[id]; !ok {
		return nil, apperrors.NewNotFoundError("inventory", id)
	}
	f.stock[id] += *dto.Quantity
	return &inventoryModels.Inventory{Quantity: f.stock[id]}, nil
}

func (f *fakeInventory) DeleteInventory(context.Context, string) error { return nil }

func (f *fakeInventory) GetInventoryByID(_ context.Context, id string) (*inventoryModels.Inventory, error) {
	qty, ok := f.stock[id]
	if !ok {
		return nil, apperrors.NewNotFoundError("inventory", id)
	}
	return &inventoryModels.Inventory{Quantity: qty}, nil
}

type fakeProducts struct {
	prices map[string]float64
}

func (f *fakeProducts) CreateProduct(context.Context, *productDTOs.CreateProductRequest) (*productDTOs.CreateProductResponse, error) {
	return nil, nil
}

func (f *fakeProducts) GetProductByID(_ context.Context, id string) (*productModels.Product, error) {
	price, ok := f.prices[id]
	if !ok {
		return nil, apperrors.NewNotFoundError("product", id)
	}
	return &productModels.Product{Price: price}, nil
}

func (f *fakeProducts) UpdateProduct(context.Context, string, *productDTOs.UpdateProductRequest) (*productModels.Product, error) {
	return nil, nil
}

func (f *fakeProducts) DeleteProduct(context.Context, string) error { return nil }

// --- helpers -------------------------------------------------------------

func newTestService(repo *fakeOrderRepo, inv *fakeInventory, products *fakeProducts) *OrderService {
	return NewOrderService(repo, inv, products, validation.NewValidator())
}

func orderRequest(items ...dtos.CreateOrderItemRequest) dtos.CreateOrderRequest {
	return dtos.CreateOrderRequest{
		CustomerID:      primitive.NewObjectID(),
		Items:           items,
		ShippingAddress: "1 Test Street",
	}
}

// --- CreateOrder ---------------------------------------------------------

func TestCreateOrder(t *testing.T) {
	productA := primitive.NewObjectID()
	productB := primitive.NewObjectID()

	tests := []struct {
		name        string
		stock       map[string]int
		req         dtos.CreateOrderRequest
		wantErrType apperrors.ErrorType
		wantTotal   float64
		wantStock   map[string]int
	}{
		{
			name:      "creates order, totals items and reserves stock",
			stock:     map[string]int{productA.Hex(): 5, productB.Hex(): 3},
			req:       orderRequest(dtos.CreateOrderItemRequest{ProductID: productA, Quantity: 2}, dtos.CreateOrderItemRequest{ProductID: productB, Quantity: 1}),
			wantTotal: 2*10.0 + 1*4.5,
			wantStock: map[string]int{productA.Hex(): 3, productB.Hex(): 2},
		},
		{
			name:        "rejects insufficient stock without touching inventory",
			stock:       map[string]int{productA.Hex(): 1},
			req:         orderRequest(dtos.CreateOrderItemRequest{ProductID: productA, Quantity: 2}),
			wantErrType: apperrors.ConflictType,
			wantStock:   map[string]int{productA.Hex(): 1},
		},
		{
			name:        "rejects unknown product",
			stock:       map[string]int{},
			req:         orderRequest(dtos.CreateOrderItemRequest{ProductID: primitive.NewObjectID(), Quantity: 1}),
			wantErrType: apperrors.NotFoundErrorType,
			wantStock:   map[string]int{},
		},
		{
			name:        "rejects invalid request",
			stock:       map[string]int{},
			req:         dtos.CreateOrderRequest{},
			wantErrType: apperrors.ValidationErrorType,
			wantStock:   map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeOrderRepo{}
			inv := &fakeInventory{stock: tt.stock}
			products := &fakeProducts{prices: map[string]float64{productA.Hex(): 10, productB.Hex(): 4.5}}

			order, err := newTestService(repo, inv, products).CreateOrder(context.Background(), tt.req)

			if tt.wantErrType != "" {
				if err == nil {
					t.Fatalf("expected %s error, got nil", tt.wantErrType)
				}
				if got := errorType(err); got != tt.wantErrType {
					t.Fatalf("error type = %q, want %q (err: %v)", got, tt.wantErrType, err)
				}
				if repo.created != nil {
					t.Fatal("order must not be persisted on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if order.TotalAmount != tt.wantTotal {
					t.Errorf("total = %v, want %v", order.TotalAmount, tt.wantTotal)
				}
				if order.Status != models.OrderStatusPending {
					t.Errorf("status = %q, want %q", order.Status, models.OrderStatusPending)
				}
			}

			for id, want := range tt.wantStock {
				if got := inv.stock[id]; got != want {
					t.Errorf("stock[%s] = %d, want %d", id, got, want)
				}
			}
		})
	}
}

func TestCreateOrder_RollsBackInventoryUsingHexProductID(t *testing.T) {
	product := primitive.NewObjectID()
	repo := &fakeOrderRepo{createErr: stderrors.New("db down")}
	inv := &fakeInventory{stock: map[string]int{product.Hex(): 5}}
	products := &fakeProducts{prices: map[string]float64{product.Hex(): 10}}

	_, err := newTestService(repo, inv, products).
		CreateOrder(context.Background(), orderRequest(dtos.CreateOrderItemRequest{ProductID: product, Quantity: 2}))

	if err == nil {
		t.Fatal("expected error when the repository fails")
	}

	want := []inventoryUpdate{
		{id: product.Hex(), quantity: -2}, // reserve
		{id: product.Hex(), quantity: 2},  // rollback
	}
	if len(inv.updates) != len(want) {
		t.Fatalf("inventory updates = %+v, want %+v", inv.updates, want)
	}
	for i := range want {
		if inv.updates[i] != want[i] {
			t.Errorf("update[%d] = %+v, want %+v", i, inv.updates[i], want[i])
		}
	}
	if got := inv.stock[product.Hex()]; got != 5 {
		t.Errorf("stock after rollback = %d, want 5", got)
	}
}

func TestCreateOrder_SurfacesRollbackFailure(t *testing.T) {
	product := primitive.NewObjectID()
	createErr := stderrors.New("db down")
	rollbackErr := stderrors.New("inventory unavailable")
	repo := &fakeOrderRepo{createErr: createErr}
	inv := &fakeInventory{stock: map[string]int{product.Hex(): 5}, rollbackErr: rollbackErr}
	products := &fakeProducts{prices: map[string]float64{product.Hex(): 10}}

	_, err := newTestService(repo, inv, products).
		CreateOrder(context.Background(), orderRequest(dtos.CreateOrderItemRequest{ProductID: product, Quantity: 2}))

	if !stderrors.Is(err, rollbackErr) {
		t.Errorf("error does not carry the rollback failure: %v", err)
	}
	if !stderrors.Is(err, createErr) {
		t.Errorf("error does not carry the original create failure: %v", err)
	}
}

// --- status transitions --------------------------------------------------

func TestIsValidStatusTransition(t *testing.T) {
	tests := []struct {
		from, to models.OrderStatus
		want     bool
	}{
		{models.OrderStatusPending, models.OrderStatusProcessing, true},
		{models.OrderStatusPending, models.OrderStatusCancelled, true},
		{models.OrderStatusProcessing, models.OrderStatusShipped, true},
		{models.OrderStatusProcessing, models.OrderStatusCancelled, true},
		{models.OrderStatusShipped, models.OrderStatusDelivered, true},
		{models.OrderStatusPending, models.OrderStatusShipped, false},
		{models.OrderStatusPending, models.OrderStatusDelivered, false},
		{models.OrderStatusShipped, models.OrderStatusCancelled, false},
		{models.OrderStatusDelivered, models.OrderStatusCancelled, false},
		{models.OrderStatusCancelled, models.OrderStatusPending, false},
		{models.OrderStatusProcessing, models.OrderStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			if got := isValidStatusTransition(tt.from, tt.to); got != tt.want {
				t.Errorf("isValidStatusTransition(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestUpdateOrderStatus(t *testing.T) {
	orderID := primitive.NewObjectID().Hex()

	tests := []struct {
		name        string
		current     models.OrderStatus
		next        models.OrderStatus
		wantErrType apperrors.ErrorType
	}{
		{name: "allowed transition is persisted", current: models.OrderStatusPending, next: models.OrderStatusProcessing},
		{name: "disallowed transition is rejected", current: models.OrderStatusDelivered, next: models.OrderStatusPending, wantErrType: apperrors.BadRequestType},
		{name: "unknown status is rejected", current: models.OrderStatusPending, next: "LOST", wantErrType: apperrors.BadRequestType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeOrderRepo{orders: map[string]*models.Order{orderID: {Status: tt.current}}}
			svc := newTestService(repo, &fakeInventory{}, &fakeProducts{})

			err := svc.UpdateOrderStatus(context.Background(), orderID, tt.next)

			if tt.wantErrType != "" {
				if got := errorType(err); got != tt.wantErrType {
					t.Fatalf("error type = %q, want %q (err: %v)", got, tt.wantErrType, err)
				}
				if _, written := repo.updated[orderID]; written {
					t.Fatal("rejected transition must not be persisted")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := repo.updated[orderID]; got != tt.next {
				t.Errorf("persisted status = %q, want %q", got, tt.next)
			}
		})
	}
}

func TestUpdateOrderStatus_OrderNotFound(t *testing.T) {
	svc := newTestService(&fakeOrderRepo{}, &fakeInventory{}, &fakeProducts{})

	err := svc.UpdateOrderStatus(context.Background(), primitive.NewObjectID().Hex(), models.OrderStatusProcessing)

	if got := errorType(err); got != apperrors.NotFoundErrorType {
		t.Fatalf("error type = %q, want %q (err: %v)", got, apperrors.NotFoundErrorType, err)
	}
}

func errorType(err error) apperrors.ErrorType {
	var appErr *apperrors.AppError
	if stderrors.As(err, &appErr) {
		return appErr.Type
	}
	var validationErrs apperrors.ValidationErrors
	if stderrors.As(err, &validationErrs) {
		return apperrors.ValidationErrorType
	}
	return ""
}
