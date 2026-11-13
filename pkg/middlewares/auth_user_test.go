package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devbenho/luka-platform/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func setTestEnv(t *testing.T) {
	t.Helper()
	t.Setenv("PORT", "0")
	t.Setenv("MONGO_URI", "mongodb://unused")
	t.Setenv("DB_NAME", "test")
	t.Setenv("JWT_SECRET", testSecret)
}

func signToken(t *testing.T, secret string, exp time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"payload": map[string]interface{}{"Id": "user-1", "Role": "admin"},
		"exp":     exp.Unix(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("signing test token: %v", err)
	}
	return token
}

func TestJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setTestEnv(t)

	validToken, err := tokens.GenerateAccessToken(map[string]interface{}{"Id": "user-1", "Role": "admin"})
	if err != nil {
		t.Fatalf("generating token: %v", err)
	}

	tests := []struct {
		name       string
		header     string
		wantStatus int
		wantUserID string
	}{
		{name: "missing header", header: "", wantStatus: http.StatusUnauthorized},
		{name: "malformed token", header: "Bearer not-a-jwt", wantStatus: http.StatusUnauthorized},
		{name: "wrong signing secret", header: "Bearer " + signToken(t, "other-secret", time.Now().Add(time.Hour)), wantStatus: http.StatusUnauthorized},
		{name: "expired token", header: "Bearer " + signToken(t, testSecret, time.Now().Add(-time.Hour)), wantStatus: http.StatusUnauthorized},
		{name: "valid bearer token", header: "Bearer " + validToken, wantStatus: http.StatusOK, wantUserID: "user-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotUserID interface{}
			router := gin.New()
			router.GET("/protected", JWT("access"), func(c *gin.Context) {
				gotUserID, _ = c.Get("userId")
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusOK && gotUserID != tt.wantUserID {
				t.Errorf("userId in context = %v, want %v", gotUserID, tt.wantUserID)
			}
			if tt.wantStatus != http.StatusOK && gotUserID != nil {
				t.Error("handler must not run when authentication fails")
			}
		})
	}
}
