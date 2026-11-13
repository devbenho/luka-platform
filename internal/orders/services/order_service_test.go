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
	storeModels "github.com/devbenho/luka-platform/internal/store/models"
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

func (f *fakeInventory) CreateInventory(context.Context, string, inventoryDTOs.CreateInventoryRequest) (*inventoryModels.Inventory, error) {
	return nil, nil
}

func (f *fakeInventory) UpdateInventory(context.Context, string, string, inventoryDTOs.UpdateInventoryRequest) (*inventoryModels.Inventory, error) {
	return nil, nil
}

func (f *fakeInventory) DeleteInventory(context.Context, string, string) error { return nil }

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
	prices  map[string]float64
	storeOf map[string]primitive.ObjectID // product hex -> store ID
}

func (f *fakeProducts) CreateProduct(context.Context, string, *productDTOs.CreateProductRequest) (*productDTOs.CreateProductResponse, error) {
	return nil, nil
}

func (f *fakeProducts) GetProductByID(_ context.Context, id string) (*productModels.Product, error) {
	price, ok := f.prices[id]
	if !ok {
		return nil, apperrors.NewNotFoundError("product", id)
	}
	return &productModels.Product{Price: price, StoreID: f.storeOf[id]}, nil
}

func (f *fakeProducts) UpdateProduct(context.Context, string, string, *productDTOs.UpdateProductRequest) (*productModels.Product, error) {
	return nil, nil
}

func (f *fakeProducts) DeleteProduct(context.Context, string, string) error { return nil }

// fakeStores maps store hex IDs to their owner.
type fakeStores struct {
	owners map[string]primitive.ObjectID
}

func (f *fakeStores) CreateStore(context.Context, *storeModels.Store) (*storeModels.Store, error) {
	return nil, nil
}

func (f *fakeStores) GetStoreByID(_ context.Context, id string) (*storeModels.Store, error) {
	owner, ok := f.owners[id]
	if !ok {
		return nil, nil
	}
	return &storeModels.Store{OwnerId: owner}, nil
}

func (f *fakeStores) UpdateStore(context.Context, string, *storeModels.Store) error { return nil }

func (f *fakeStores) DeleteStore(context.Context, string) error { return nil }

// --- helpers -------------------------------------------------------------

func newTestService(repo repositories.IOrderRepository, inv *fakeInventory, products *fakeProducts) *OrderService {
	return newTestServiceWithStores(repo, inv, products, &fakeStores{})
}

func newTestServiceWithStores(repo repositories.IOrderRepository, inv *fakeInventory, products *fakeProducts, stores *fakeStores) *OrderService {
	return NewOrderService(repo, inv, products, stores, &fakeTx{inv: inv, repo: repo}, validation.NewValidator())
}

// fakeTx imitates a MongoDB transaction: if fn fails, the stock levels and
// order state it changed are restored, as an aborted transaction would.
type fakeTx struct {
	inv     *fakeInventory
	repo    repositories.IOrderRepository
	runs    int
	aborted int
}

func (f *fakeTx) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	f.runs++
	stock := map[string]int{}
	for k, v := range f.inv.stock {
		stock[k] = v
	}
	orders := f.orderRepo()
	var created *models.Order
	statuses := map[string]models.OrderStatus{}
	if orders != nil {
		created = orders.created
		for id, o := range orders.orders {
			statuses[id] = o.Status
		}
	}

	err := fn(ctx)
	if err != nil {
		f.aborted++
		f.inv.stock = stock
		if orders != nil {
			orders.created = created
			for id, status := range statuses {
				orders.orders[id].Status = status
			}
			orders.updated = nil
		}
	}
	return err
}

func (f *fakeTx) orderRepo() *fakeOrderRepo {
	switch r := f.repo.(type) {
	case *fakeOrderRepo:
		return r
	case *stalePendingRepo:
		return r.fakeOrderRepo
	}
	return nil
}

// marketplace is one seller's store with one product, and one customer.
type marketplace struct {
	seller, customer, stranger string
	product                    primitive.ObjectID
	products                   *fakeProducts
	stores                     *fakeStores
}

func newMarketplace() marketplace {
	seller, store, product := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	return marketplace{
		seller:   seller.Hex(),
		customer: primitive.NewObjectID().Hex(),
		stranger: primitive.NewObjectID().Hex(),
		product:  product,
		products: &fakeProducts{
			prices:  map[string]float64{product.Hex(): 10},
			storeOf: map[string]primitive.ObjectID{product.Hex(): store},
		},
		stores: &fakeStores{owners: map[string]primitive.ObjectID{store.Hex(): seller}},
	}
}

func (m marketplace) order(status models.OrderStatus) *models.Order {
	customer, _ := primitive.ObjectIDFromHex(m.customer)
	return &models.Order{
		CustomerID: customer,
		Status:     status,
		Items:      []models.OrderItem{{ProductID: m.product, Quantity: 1}},
	}
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

func TestCreateOrder_SaveFailureAbortsTheTransaction(t *testing.T) {
	product := primitive.NewObjectID()
	saveErr := stderrors.New("db down")
	repo := &fakeOrderRepo{createErr: saveErr}
	inv := newFakeInventory(map[string]int{product.Hex(): 5})
	products := &fakeProducts{prices: map[string]float64{product.Hex(): 10}}
	tx := &fakeTx{inv: inv, repo: repo}
	svc := NewOrderService(repo, inv, products, &fakeStores{}, tx, validation.NewValidator())

	_, err := svc.CreateOrder(context.Background(), orderRequest(dtos.CreateOrderItemRequest{ProductID: product, Quantity: 2}))

	if !stderrors.Is(err, saveErr) {
		t.Fatalf("error = %v, want it to wrap the save failure", err)
	}
	if want := []stockMove{{id: product.Hex(), quantity: 2}}; !equalMoves(inv.reserves, want) {
		t.Errorf("reserves = %+v, want %+v", inv.reserves, want)
	}
	if tx.runs != 1 || tx.aborted != 1 {
		t.Errorf("transaction runs/aborts = %d/%d, want the reservation and save in one aborted transaction", tx.runs, tx.aborted)
	}
	if got := inv.stock[product.Hex()]; got != 5 {
		t.Errorf("stock after abort = %d, want 5", got)
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
	m := newMarketplace()
	orderID := primitive.NewObjectID().Hex()

	tests := []struct {
		name        string
		actor       string
		current     models.OrderStatus
		next        models.OrderStatus
		wantErrType apperrors.ErrorType
	}{
		{name: "seller advances the order", actor: m.seller, current: models.OrderStatusPending, next: models.OrderStatusProcessing},
		{name: "seller ships the order", actor: m.seller, current: models.OrderStatusProcessing, next: models.OrderStatusShipped},
		{name: "seller cancels the order", actor: m.seller, current: models.OrderStatusPending, next: models.OrderStatusCancelled},
		{name: "customer cancels own order", actor: m.customer, current: models.OrderStatusPending, next: models.OrderStatusCancelled},
		{name: "customer cannot ship", actor: m.customer, current: models.OrderStatusProcessing, next: models.OrderStatusShipped, wantErrType: apperrors.ForbiddenType},
		{name: "stranger cannot cancel", actor: m.stranger, current: models.OrderStatusPending, next: models.OrderStatusCancelled, wantErrType: apperrors.ForbiddenType},
		{name: "disallowed transition is rejected", actor: m.seller, current: models.OrderStatusDelivered, next: models.OrderStatusPending, wantErrType: apperrors.BadRequestType},
		{name: "customer cannot cancel a shipped order", actor: m.customer, current: models.OrderStatusShipped, next: models.OrderStatusCancelled, wantErrType: apperrors.BadRequestType},
		{name: "unknown status is rejected", actor: m.seller, current: models.OrderStatusPending, next: "LOST", wantErrType: apperrors.BadRequestType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeOrderRepo{orders: map[string]*models.Order{orderID: m.order(tt.current)}}
			svc := newTestServiceWithStores(repo, newFakeInventory(nil), m.products, m.stores)

			err := svc.UpdateOrderStatus(context.Background(), orderID, tt.actor, tt.next)

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

func TestGetOrderByID_Access(t *testing.T) {
	m := newMarketplace()
	orderID := primitive.NewObjectID().Hex()

	tests := []struct {
		name        string
		actor       string
		wantErrType apperrors.ErrorType
	}{
		{name: "customer sees own order", actor: m.customer},
		{name: "seller sees order for their products", actor: m.seller},
		{name: "stranger is forbidden", actor: m.stranger, wantErrType: apperrors.ForbiddenType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeOrderRepo{orders: map[string]*models.Order{orderID: m.order(models.OrderStatusPending)}}
			svc := newTestServiceWithStores(repo, newFakeInventory(nil), m.products, m.stores)

			order, err := svc.GetOrderByID(context.Background(), orderID, tt.actor)

			if tt.wantErrType != "" {
				if got := errorType(err); got != tt.wantErrType {
					t.Fatalf("error type = %q, want %q (err: %v)", got, tt.wantErrType, err)
				}
				return
			}
			if err != nil || order == nil {
				t.Fatalf("expected the order, got %v / %v", order, err)
			}
		})
	}
}

func TestUpdateOrderStatus_CancelReleasesStock(t *testing.T) {
	m := newMarketplace()
	inv := newFakeInventory(map[string]int{m.product.Hex(): 3})
	inventoryID := inv.inventoryOf[m.product.Hex()]
	order := m.order(models.OrderStatusPending)
	order.Items[0].Quantity = 2
	order.Items[0].InventoryID = inventoryID
	orderID := primitive.NewObjectID().Hex()
	repo := &fakeOrderRepo{orders: map[string]*models.Order{orderID: order}}

	err := newTestServiceWithStores(repo, inv, m.products, m.stores).UpdateOrderStatus(context.Background(), orderID, m.customer, models.OrderStatusCancelled)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []stockMove{{id: inventoryID.Hex(), quantity: 2}}; !equalMoves(inv.releases, want) {
		t.Errorf("releases = %+v, want %+v", inv.releases, want)
	}
	if got := inv.stock[m.product.Hex()]; got != 5 {
		t.Errorf("stock after cancel = %d, want 5", got)
	}
}

func TestUpdateOrderStatus_ConcurrentChangeIsRejected(t *testing.T) {
	// The request read the order while it was PENDING, but another request
	// has since moved it to PROCESSING. The guarded write must fail rather
	// than cancel the order (and release its stock) a second time.
	m := newMarketplace()
	orderID := primitive.NewObjectID().Hex()
	repo := &stalePendingRepo{&fakeOrderRepo{orders: map[string]*models.Order{orderID: m.order(models.OrderStatusProcessing)}}, m.order(models.OrderStatusPending)}
	inv := newFakeInventory(map[string]int{})

	err := newTestServiceWithStores(repo, inv, m.products, m.stores).UpdateOrderStatus(context.Background(), orderID, m.customer, models.OrderStatusCancelled)

	if got := errorType(err); got != apperrors.ConflictType {
		t.Fatalf("error type = %q, want %q (err: %v)", got, apperrors.ConflictType, err)
	}
	if len(inv.releases) != 0 {
		t.Errorf("stock released despite the conflict: %+v", inv.releases)
	}
}

func TestUpdateOrderStatus_CancelIsUndoneWhenStockCannotBeReleased(t *testing.T) {
	m := newMarketplace()
	inv := newFakeInventory(map[string]int{m.product.Hex(): 3})
	inv.releaseErr = stderrors.New("inventory unavailable")
	order := m.order(models.OrderStatusPending)
	order.Items[0].InventoryID = inv.inventoryOf[m.product.Hex()]
	orderID := primitive.NewObjectID().Hex()
	repo := &fakeOrderRepo{orders: map[string]*models.Order{orderID: order}}

	err := newTestServiceWithStores(repo, inv, m.products, m.stores).UpdateOrderStatus(context.Background(), orderID, m.customer, models.OrderStatusCancelled)

	if !stderrors.Is(err, inv.releaseErr) {
		t.Fatalf("error = %v, want the release failure", err)
	}
	if got := repo.orders[orderID].Status; got != models.OrderStatusPending {
		t.Errorf("status after failed cancel = %q, want it unchanged (%q)", got, models.OrderStatusPending)
	}
}

// stalePendingRepo returns an out-of-date snapshot while the stored order has
// already moved on, simulating a request that lost a race.
type stalePendingRepo struct {
	*fakeOrderRepo
	snapshot *models.Order
}

func (r *stalePendingRepo) GetOrderByID(context.Context, string) (*models.Order, error) {
	return r.snapshot, nil
}

func TestUpdateOrderStatus_OrderNotFound(t *testing.T) {
	svc := newTestService(&fakeOrderRepo{}, newFakeInventory(nil), &fakeProducts{})

	err := svc.UpdateOrderStatus(context.Background(), primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex(), models.OrderStatusProcessing)

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
