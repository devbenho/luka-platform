package services

import (
	"context"
	stderrors "errors"
	"testing"

	inventoryDTOs "github.com/devbenho/luka-platform/internal/inventory/dtos"
	inventoryModels "github.com/devbenho/luka-platform/internal/inventory/models"
	"github.com/devbenho/luka-platform/internal/orders/models"
	dtos "github.com/devbenho/luka-platform/internal/orders/order_dtos"
	"github.com/devbenho/luka-platform/internal/orders/repositories"
	productDTOs "github.com/devbenho/luka-platform/internal/product/dtos"
	productModels "github.com/devbenho/luka-platform/internal/product/models"
	apperrors "github.com/devbenho/luka-platform/pkg/errors"
	"github.com/devbenho/luka-platform/pkg/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	return r.UpdateOrderStatus(context.Background(), id, r.orders[id].Status, order.Status)
}

func (r *fakeOrderRepo) UpdateOrderStatus(_ context.Context, id string, from, to models.OrderStatus) error {
	order, ok := r.orders[id]
	if !ok || order.Status != from {
		return mongo.ErrNoDocuments
	}
	if r.updated == nil {
		r.updated = map[string]models.OrderStatus{}
	}
	order.Status = to
	r.updated[id] = to
	return nil
}

func (r *fakeOrderRepo) GetOrderByID(_ context.Context, id string) (*models.Order, error) {
	return r.orders[id], nil
}

func (r *fakeOrderRepo) ListOrders(context.Context, string) ([]models.Order, error) {
	return nil, nil
}

type stockMove struct {
	id       string // product hex for reserves, inventory hex for releases
	quantity int
}

// fakeInventory holds one inventory record per product. It records every
// reserve (by product ID) and release (by inventory ID) so tests can assert
// that stock is released to the same record it was reserved from.
type fakeInventory struct {
	stock       map[string]int                // product hex -> quantity
	inventoryOf map[string]primitive.ObjectID // product hex -> inventory ID
	reserves    []stockMove
	releases    []stockMove
	releaseErr  error
}

func newFakeInventory(stock map[string]int) *fakeInventory {
	f := &fakeInventory{stock: stock, inventoryOf: map[string]primitive.ObjectID{}}
	for product := range stock {
		f.inventoryOf[product] = primitive.NewObjectID()
	}
	return f
}

func (f *fakeInventory) CreateInventory(context.Context, inventoryDTOs.CreateInventoryRequest) (*inventoryModels.Inventory, error) {
	return nil, nil
}

func (f *fakeInventory) UpdateInventory(context.Context, string, inventoryDTOs.UpdateInventoryRequest) (*inventoryModels.Inventory, error) {
	return nil, nil
}

func (f *fakeInventory) DeleteInventory(context.Context, string) error { return nil }

func (f *fakeInventory) GetInventoryByID(context.Context, string) (*inventoryModels.Inventory, error) {
	return nil, nil
}

func (f *fakeInventory) ReserveStock(_ context.Context, productID primitive.ObjectID, quantity int) (primitive.ObjectID, error) {
	product := productID.Hex()
	f.reserves = append(f.reserves, stockMove{id: product, quantity: quantity})
	if f.stock[product] < quantity {
		return primitive.NilObjectID, apperrors.NewConflictError("insufficient stock")
	}
	f.stock[product] -= quantity
	return f.inventoryOf[product], nil
}

func (f *fakeInventory) ReleaseStock(_ context.Context, inventoryID primitive.ObjectID, quantity int) error {
	f.releases = append(f.releases, stockMove{id: inventoryID.Hex(), quantity: quantity})
	if f.releaseErr != nil {
		return f.releaseErr
	}
	for product, id := range f.inventoryOf {
		if id == inventoryID {
			f.stock[product] += quantity
			return nil
		}
	}
	return mongo.ErrNoDocuments
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

func newTestService(repo repositories.IOrderRepository, inv *fakeInventory, products *fakeProducts) *OrderService {
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
			name:        "rejects insufficient stock",
			stock:       map[string]int{productA.Hex(): 1},
			req:         orderRequest(dtos.CreateOrderItemRequest{ProductID: productA, Quantity: 2}),
			wantErrType: apperrors.ConflictType,
			wantStock:   map[string]int{productA.Hex(): 1},
		},
		{
			name:        "releases earlier items when a later item is out of stock",
			stock:       map[string]int{productA.Hex(): 5, productB.Hex(): 0},
			req:         orderRequest(dtos.CreateOrderItemRequest{ProductID: productA, Quantity: 2}, dtos.CreateOrderItemRequest{ProductID: productB, Quantity: 1}),
			wantErrType: apperrors.ConflictType,
			wantStock:   map[string]int{productA.Hex(): 5, productB.Hex(): 0},
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
			inv := newFakeInventory(tt.stock)
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
				for _, item := range order.Items {
					if item.InventoryID != inv.inventoryOf[item.ProductID.Hex()] {
						t.Errorf("item %s inventory = %s, want the record it was reserved from", item.ProductID.Hex(), item.InventoryID.Hex())
					}
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

func TestCreateOrder_ReleasesReservedStockWhenSavingFails(t *testing.T) {
	product := primitive.NewObjectID()
	repo := &fakeOrderRepo{createErr: stderrors.New("db down")}
	inv := newFakeInventory(map[string]int{product.Hex(): 5})
	products := &fakeProducts{prices: map[string]float64{product.Hex(): 10}}

	_, err := newTestService(repo, inv, products).
		CreateOrder(context.Background(), orderRequest(dtos.CreateOrderItemRequest{ProductID: product, Quantity: 2}))

	if err == nil {
		t.Fatal("expected error when the repository fails")
	}
	wantReserves := []stockMove{{id: product.Hex(), quantity: 2}}
	wantReleases := []stockMove{{id: inv.inventoryOf[product.Hex()].Hex(), quantity: 2}}
	if !equalMoves(inv.reserves, wantReserves) {
		t.Errorf("reserves = %+v, want %+v", inv.reserves, wantReserves)
	}
	if !equalMoves(inv.releases, wantReleases) {
		t.Errorf("releases = %+v, want %+v (must target the reserved inventory record)", inv.releases, wantReleases)
	}
	if got := inv.stock[product.Hex()]; got != 5 {
		t.Errorf("stock after rollback = %d, want 5", got)
	}
}

func TestCreateOrder_SurfacesRollbackFailure(t *testing.T) {
	product := primitive.NewObjectID()
	createErr := stderrors.New("db down")
	releaseErr := stderrors.New("inventory unavailable")
	repo := &fakeOrderRepo{createErr: createErr}
	inv := newFakeInventory(map[string]int{product.Hex(): 5})
	inv.releaseErr = releaseErr
	products := &fakeProducts{prices: map[string]float64{product.Hex(): 10}}

	_, err := newTestService(repo, inv, products).
		CreateOrder(context.Background(), orderRequest(dtos.CreateOrderItemRequest{ProductID: product, Quantity: 2}))

	if !stderrors.Is(err, releaseErr) {
		t.Errorf("error does not carry the rollback failure: %v", err)
	}
	if !stderrors.Is(err, createErr) {
		t.Errorf("error does not carry the original create failure: %v", err)
	}
}

func equalMoves(got, want []stockMove) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
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
			svc := newTestService(repo, newFakeInventory(nil), &fakeProducts{})

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

func TestUpdateOrderStatus_CancelReleasesStock(t *testing.T) {
	product := primitive.NewObjectID()
	inv := newFakeInventory(map[string]int{product.Hex(): 3})
	inventoryID := inv.inventoryOf[product.Hex()]
	orderID := primitive.NewObjectID().Hex()
	repo := &fakeOrderRepo{orders: map[string]*models.Order{orderID: {
		Status: models.OrderStatusPending,
		Items:  []models.OrderItem{{ProductID: product, Quantity: 2, InventoryID: inventoryID}},
	}}}

	err := newTestService(repo, inv, &fakeProducts{}).UpdateOrderStatus(context.Background(), orderID, models.OrderStatusCancelled)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []stockMove{{id: inventoryID.Hex(), quantity: 2}}; !equalMoves(inv.releases, want) {
		t.Errorf("releases = %+v, want %+v", inv.releases, want)
	}
	if got := inv.stock[product.Hex()]; got != 5 {
		t.Errorf("stock after cancel = %d, want 5", got)
	}
}

func TestUpdateOrderStatus_ConcurrentChangeIsRejected(t *testing.T) {
	// The request read the order while it was PENDING, but another request
	// has since moved it to PROCESSING. The guarded write must fail rather
	// than cancel the order (and release its stock) a second time.
	orderID := primitive.NewObjectID().Hex()
	repo := &stalePendingRepo{&fakeOrderRepo{orders: map[string]*models.Order{orderID: {Status: models.OrderStatusProcessing}}}}
	inv := newFakeInventory(map[string]int{})

	err := newTestService(repo, inv, &fakeProducts{}).UpdateOrderStatus(context.Background(), orderID, models.OrderStatusCancelled)

	if got := errorType(err); got != apperrors.ConflictType {
		t.Fatalf("error type = %q, want %q (err: %v)", got, apperrors.ConflictType, err)
	}
	if len(inv.releases) != 0 {
		t.Errorf("stock released despite the conflict: %+v", inv.releases)
	}
}

// stalePendingRepo returns a PENDING snapshot while the stored order has
// already moved on, simulating a request that lost a race.
type stalePendingRepo struct{ *fakeOrderRepo }

func (r *stalePendingRepo) GetOrderByID(context.Context, string) (*models.Order, error) {
	return &models.Order{Status: models.OrderStatusPending}, nil
}

func TestUpdateOrderStatus_OrderNotFound(t *testing.T) {
	svc := newTestService(&fakeOrderRepo{}, newFakeInventory(nil), &fakeProducts{})

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
