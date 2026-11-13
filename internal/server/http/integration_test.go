//go:build integration

// Integration tests run the real HTTP routes against a real MongoDB.
//
//	LUKA_TEST_MONGO_URI="mongodb://localhost:27017" go test -tags integration ./internal/server/http/
//
// Each run uses a fresh, randomly named database that is dropped afterwards.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/devbenho/luka-platform/pkg/database"
	"github.com/devbenho/luka-platform/pkg/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type apiClient struct {
	t      *testing.T
	engine *gin.Engine
}

type apiResponse struct {
	Status int
	Body   map[string]interface{}
}

func (r apiResponse) data() map[string]interface{} {
	d, _ := r.Body["data"].(map[string]interface{})
	return d
}

func newAPI(t *testing.T) *apiClient {
	t.Helper()
	uri := os.Getenv("LUKA_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("LUKA_TEST_MONGO_URI not set")
	}
	dbName := "luka_it_" + primitive.NewObjectID().Hex()

	t.Setenv("PORT", "0")
	t.Setenv("MONGO_URI", uri)
	t.Setenv("DB_NAME", dbName)
	t.Setenv("JWT_SECRET", "integration-secret")

	db, err := database.NewDatabase(uri, dbName)
	if err != nil {
		t.Fatalf("connecting to mongo: %v", err)
	}
	t.Cleanup(func() { _ = db.GetDB().Drop(context.Background()) })

	gin.SetMode(gin.TestMode)
	server := NewServer(validation.NewValidator(), db, zap.NewNop())
	if err := server.MapRoutes(); err != nil {
		t.Fatalf("mapping routes: %v", err)
	}
	return &apiClient{t: t, engine: server.GetEngine()}
}

func (a *apiClient) do(method, path, token string, body interface{}) apiResponse {
	a.t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			a.t.Fatalf("encoding body: %v", err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, "/api/v1"+path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	a.engine.ServeHTTP(rec, req)

	resp := apiResponse{Status: rec.Code}
	if rec.Body.Len() > 0 {
		_ = json.Unmarshal(rec.Body.Bytes(), &resp.Body)
	}
	return resp
}

func (a *apiClient) expect(resp apiResponse, status int, what string) apiResponse {
	a.t.Helper()
	if resp.Status != status {
		a.t.Fatalf("%s: status = %d, want %d; body = %v", what, resp.Status, status, resp.Body)
	}
	return resp
}

type testUser struct {
	ID       string
	Token    string
	Username string
}

func (a *apiClient) register(username, role string) testUser {
	a.t.Helper()
	resp := a.expect(a.do(http.MethodPost, "/auth/register", "", map[string]string{
		"username": username,
		"email":    username + "@example.com",
		"password": "s3cret-pass",
		"role":     role,
	}), http.StatusCreated, "register "+username)
	return testUser{ID: resp.data()["id"].(string), Token: resp.data()["token"].(string), Username: username}
}

func TestUsersAndAuth(t *testing.T) {
	api := newAPI(t)
	alice := api.register("alice", "seller")
	bob := api.register("bob", "buyer")

	api.expect(api.do(http.MethodPost, "/auth/register", "", map[string]string{
		"username": "alice", "email": "other@example.com", "password": "s3cret-pass", "role": "buyer",
	}), http.StatusBadRequest, "duplicate username")

	for _, login := range []string{"alice", "alice@example.com"} {
		resp := api.expect(api.do(http.MethodPost, "/auth/login", "", map[string]string{
			"login": login, "password": "s3cret-pass",
		}), http.StatusOK, "login as "+login)
		if resp.data()["token"] == "" {
			t.Errorf("login as %s returned no token", login)
		}
	}

	wrongPassword := api.expect(api.do(http.MethodPost, "/auth/login", "", map[string]string{
		"login": "alice", "password": "wrong-pass",
	}), http.StatusUnauthorized, "wrong password")
	unknownUser := api.expect(api.do(http.MethodPost, "/auth/login", "", map[string]string{
		"login": "nobody", "password": "wrong-pass",
	}), http.StatusUnauthorized, "unknown user")
	if wrongPassword.Body["message"] != unknownUser.Body["message"] || wrongPassword.Body["error"] != unknownUser.Body["error"] {
		t.Errorf("login errors reveal whether a user exists: %v vs %v", wrongPassword.Body, unknownUser.Body)
	}

	api.expect(api.do(http.MethodGet, "/users/"+alice.ID, alice.Token, nil), http.StatusOK, "get user")
	api.expect(api.do(http.MethodGet, "/users/not-an-id", alice.Token, nil), http.StatusBadRequest, "get user with invalid id")
	api.expect(api.do(http.MethodGet, "/users/"+primitive.NewObjectID().Hex(), alice.Token, nil), http.StatusNotFound, "get missing user")
	api.expect(api.do(http.MethodGet, "/users/"+alice.ID, "", nil), http.StatusUnauthorized, "get user without token")

	api.expect(api.do(http.MethodPut, "/users/"+alice.ID, alice.Token, map[string]string{"username": "alice2"}), http.StatusOK, "update own user")
	got := api.expect(api.do(http.MethodGet, "/users/"+alice.ID, alice.Token, nil), http.StatusOK, "get updated user")
	if got.Body["username"] != "alice2" {
		t.Errorf("username after update = %v, want alice2", got.Body["username"])
	}
	api.expect(api.do(http.MethodPut, "/users/"+alice.ID, bob.Token, map[string]string{"username": "hijack"}), http.StatusForbidden, "update another user")
	api.expect(api.do(http.MethodPut, "/users/"+alice.ID, alice.Token, map[string]string{"email": "not-an-email"}), http.StatusBadRequest, "update with invalid email")

	api.expect(api.do(http.MethodDelete, "/users/"+bob.ID, bob.Token, nil), http.StatusOK, "delete own user")
	api.expect(api.do(http.MethodGet, "/users/"+bob.ID, alice.Token, nil), http.StatusNotFound, "get deleted user")
}

func TestStores(t *testing.T) {
	api := newAPI(t)
	alice := api.register("alice", "seller")
	bob := api.register("bob", "seller")

	created := api.expect(api.do(http.MethodPost, "/stores/", alice.Token, map[string]interface{}{
		"name":       "Alice Books",
		"location":   map[string]float64{"latitude": 30, "longitude": 31},
		"store_type": map[string]bool{"online": true},
	}), http.StatusCreated, "create store")
	storeID := created.data()["id"].(string)

	got := api.expect(api.do(http.MethodGet, "/stores/"+storeID, alice.Token, nil), http.StatusOK, "get store")
	if got.data()["ownerId"] != alice.ID {
		t.Errorf("store owner = %v, want %s", got.data()["ownerId"], alice.ID)
	}

	api.expect(api.do(http.MethodPatch, "/stores/"+storeID, bob.Token, map[string]string{"name": "Bob's now"}), http.StatusForbidden, "update another user's store")
	api.expect(api.do(http.MethodDelete, "/stores/"+storeID, bob.Token, nil), http.StatusForbidden, "delete another user's store")

	api.expect(api.do(http.MethodPatch, "/stores/"+storeID, alice.Token, map[string]string{"name": "Alice Rare Books"}), http.StatusOK, "update own store")
	got = api.expect(api.do(http.MethodGet, "/stores/"+storeID, alice.Token, nil), http.StatusOK, "get updated store")
	if got.data()["name"] != "Alice Rare Books" || got.data()["slug"] != "alice-rare-books" {
		t.Errorf("store after update = %v", got.data())
	}
	if got.data()["ownerId"] != alice.ID {
		t.Errorf("update changed the store owner to %v", got.data()["ownerId"])
	}

	api.expect(api.do(http.MethodGet, "/stores/"+primitive.NewObjectID().Hex(), alice.Token, nil), http.StatusNotFound, "get missing store")
	api.expect(api.do(http.MethodGet, "/stores/bad-id", alice.Token, nil), http.StatusBadRequest, "get store with invalid id")

	api.expect(api.do(http.MethodDelete, "/stores/"+storeID, alice.Token, nil), http.StatusOK, "delete own store")
	api.expect(api.do(http.MethodGet, "/stores/"+storeID, alice.Token, nil), http.StatusNotFound, "get deleted store")
	api.expect(api.do(http.MethodDelete, "/stores/"+storeID, alice.Token, nil), http.StatusNotFound, "delete store twice")
}

func TestCategories(t *testing.T) {
	api := newAPI(t)
	alice := api.register("alice", "seller")

	created := api.expect(api.do(http.MethodPost, "/categories/", alice.Token, map[string]string{
		"name": "Books", "slug": "books", "description": "Printed books",
	}), http.StatusCreated, "create category")
	categoryID := created.data()["category"].(map[string]interface{})["id"].(string)

	if slug := created.data()["category"].(map[string]interface{})["slug"]; slug != "books" {
		t.Errorf("generated slug = %v, want books", slug)
	}
	api.expect(api.do(http.MethodPost, "/categories/", alice.Token, map[string]string{
		"name": "ab",
	}), http.StatusBadRequest, "create category with too-short name")

	api.expect(api.do(http.MethodPatch, "/categories/"+categoryID, alice.Token, map[string]string{"name": "Novels"}), http.StatusOK, "partial category update")
	got := api.expect(api.do(http.MethodGet, "/categories/"+categoryID, alice.Token, nil), http.StatusOK, "get category")
	if got.data()["name"] != "Novels" || got.data()["slug"] != "novels" || got.data()["description"] != "Printed books" {
		t.Errorf("category after partial update = %v", got.data())
	}

	api.expect(api.do(http.MethodDelete, "/categories/"+categoryID, alice.Token, nil), http.StatusOK, "delete category")
	api.expect(api.do(http.MethodGet, "/categories/"+categoryID, alice.Token, nil), http.StatusNotFound, "get deleted category")
}

func createStore(t *testing.T, api *apiClient, owner testUser, name string) string {
	t.Helper()
	resp := api.expect(api.do(http.MethodPost, "/stores/", owner.Token, map[string]interface{}{
		"name":       name,
		"store_type": map[string]bool{"online": true},
	}), http.StatusCreated, "create store "+name)
	return resp.data()["id"].(string)
}

func createProduct(t *testing.T, api *apiClient, owner testUser, storeID string, price float64) string {
	t.Helper()
	resp := api.expect(api.do(http.MethodPost, "/products/", owner.Token, map[string]interface{}{
		"name":        fmt.Sprintf("Product %v", price),
		"description": "A product",
		"price":       price,
		"store_id":    storeID,
		"categories":  []string{},
		"images":      []string{"https://example.com/p.png"},
	}), http.StatusCreated, "create product")
	return resp.data()["product"].(map[string]interface{})["id"].(string)
}

func TestProducts(t *testing.T) {
	api := newAPI(t)
	alice := api.register("alice", "seller")
	storeID := createStore(t, api, alice, "Alice Books")
	productID := createProduct(t, api, alice, storeID, 10)

	api.expect(api.do(http.MethodPost, "/products/", alice.Token, map[string]interface{}{
		"name": "Orphan", "description": "x", "price": 1, "store_id": primitive.NewObjectID().Hex(),
		"categories": []string{}, "images": []string{"x"},
	}), http.StatusNotFound, "create product in missing store")

	api.expect(api.do(http.MethodPatch, "/products/"+productID, alice.Token, map[string]interface{}{"price": 12.5}), http.StatusOK, "update product price")
	got := api.expect(api.do(http.MethodGet, "/products/"+productID, alice.Token, nil), http.StatusOK, "get product")
	if got.data()["price"] != 12.5 || got.data()["name"] != "Product 10" {
		t.Errorf("product after update = %v", got.data())
	}

	api.expect(api.do(http.MethodGet, "/products/"+primitive.NewObjectID().Hex(), alice.Token, nil), http.StatusNotFound, "get missing product")
	api.expect(api.do(http.MethodDelete, "/products/"+productID, alice.Token, nil), http.StatusOK, "delete product")
	api.expect(api.do(http.MethodGet, "/products/"+productID, alice.Token, nil), http.StatusNotFound, "get deleted product")
}

func TestOrdersList(t *testing.T) {
	api := newAPI(t)
	alice := api.register("alice", "buyer")

	resp := api.expect(api.do(http.MethodGet, "/orders/?customer_id="+alice.ID, alice.Token, nil), http.StatusOK, "list orders")
	if orders, ok := resp.Body["data"].([]interface{}); ok && len(orders) != 0 {
		t.Errorf("new user has orders: %v", orders)
	}
	api.expect(api.do(http.MethodGet, "/orders/"+primitive.NewObjectID().Hex(), alice.Token, nil), http.StatusNotFound, "get missing order")
}
