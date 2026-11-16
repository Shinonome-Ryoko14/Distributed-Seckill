package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/jwt"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/middleware"
)

func newAuthTestRouter() (*gin.Engine, *jwt.Manager) {
	gin.SetMode(gin.TestMode)
	manager := jwt.NewManager("test-secret", time.Hour, "test-issuer")
	r := NewRouter(NewProductHandler(&fakeProductService{createID: 1}),
		NewActivityHandler(&fakeActivityService{createID: 2}),
		middleware.Auth(manager))
	return r, manager
}

func doAuthedRequest(
	t *testing.T, r *gin.Engine, method, path, body, token string,
) (int, apiResponse) {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(),
		method, path,
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return w.Code, resp
}

func TestCreateRoutesRequireAuth(t *testing.T) {
	r, _ := newAuthTestRouter()

	status, resp := doAuthedRequest(t, r,
		http.MethodPost, "/api/v1/products", `{"name":"x","price":1,"status":1}`, "")
	if status != http.StatusUnauthorized || resp.Code != errs.CodeUnauthorized {
		t.Fatalf("products: status=%d code=%d, want 401/%d", status, resp.Code, errs.CodeUnauthorized)
	}

	status, resp = doAuthedRequest(t, r,
		http.MethodPost, "/api/v1/activities", validCreateActivityBody, "")
	if status != http.StatusUnauthorized || resp.Code != errs.CodeUnauthorized {
		t.Fatalf("activities: status=%d code=%d, want 401/%d", status, resp.Code, errs.CodeUnauthorized)
	}
}

func TestCreateWithValidToken(t *testing.T) {
	r, manager := newAuthTestRouter()
	token, err := manager.Generate(42)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	status, resp := doAuthedRequest(t, r, http.MethodPost,
		"/api/v1/products", `{"name":"x","price":1,"status":1}`, token)
	if status != http.StatusOK || resp.Code != errs.CodeOK {
		t.Fatalf("status=%d code=%d, want 200/0", status, resp.Code)
	}
}

func TestGetRemainsPublic(t *testing.T) {
	r, _ := newAuthTestRouter()

	status, _ := doAuthedRequest(t, r, http.MethodGet, "/api/v1/products/1", "", "")
	if status == http.StatusUnauthorized {
		t.Fatal("GET /products/1 must not require auth")
	}
}
