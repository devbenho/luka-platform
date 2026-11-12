package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOwnerAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		authUserID string // empty means no authenticated user in context
		route      string
		path       string
		wantStatus int
	}{
		{name: "owner can access own resource", authUserID: "user-1", route: "/users/:id", path: "/users/user-1", wantStatus: http.StatusOK},
		{name: "other user is forbidden", authUserID: "user-2", route: "/users/:id", path: "/users/user-1", wantStatus: http.StatusForbidden},
		{name: "unauthenticated request is rejected", route: "/users/:id", path: "/users/user-1", wantStatus: http.StatusUnauthorized},
		{name: "missing resource id is a bad request", authUserID: "user-1", route: "/users", path: "/users", wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerRan := false
			router := gin.New()
			setUser := func(c *gin.Context) {
				if tt.authUserID != "" {
					c.Set("userId", tt.authUserID)
				}
			}
			router.PUT(tt.route, setUser, OwnerAuth(), func(c *gin.Context) {
				handlerRan = true
				c.Status(http.StatusOK)
			})

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, tt.path, nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if handlerRan != (tt.wantStatus == http.StatusOK) {
				t.Errorf("handler ran = %v, want %v", handlerRan, tt.wantStatus == http.StatusOK)
			}
		})
	}
}
