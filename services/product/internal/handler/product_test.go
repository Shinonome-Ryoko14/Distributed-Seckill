package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/service"
)

type fakeProductService struct {
	createID    uint64
	createErr   error
	getResult   *model.Product
	getErr      error
	listResult  *service.ProductPage
	listErr     error
	createCalls int
}

var _ productService = (*fakeProductService)(nil)

func (f *fakeProductService) Create(_ context.Context, _ service.CreateProductInput) (uint64, error) {
	f.createCalls++
	return f.createID, f.createErr
}

func (f *fakeProductService) GetByID(_ context.Context, _ uint64) (*model.Product, error) {
	return f.getResult, f.getErr
}

func (f *fakeProductService) List(_ context.Context, _, _ int) (*service.ProductPage, error) {
	return f.listResult, f.listErr
}

type apiResponse struct {
	Code    errs.Code       `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func newTestRouter(fake *fakeProductService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return NewRouter(NewProductHandler(fake))
}

func doRequest(t *testing.T, r *gin.Engine, method, path, body string) (int, apiResponse) {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return w.Code, resp
}

func TestCreateSuccess(t *testing.T) {
	fake := &fakeProductService{createID: 7}
	r := newTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodPost, "/api/v1/products",
		`{"name":"iPhone","description":"demo","price":100,"status":1}`)

	if status != http.StatusOK || resp.Code != errs.CodeOK {
		t.Fatalf("status=%d code=%d, want 200/0", status, resp.Code)
	}
	var data struct {
		ID uint64 `json:"id"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if data.ID != 7 {
		t.Errorf("data.id = %d, want 7", data.ID)
	}
}

func TestCreateInvalidJSON(t *testing.T) {
	fake := &fakeProductService{}
	r := newTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodPost, "/api/v1/products", `{`)
	if status != http.StatusBadRequest || resp.Code != errs.CodeInvalidParam {
		t.Fatalf("status=%d code=%d, want 400/%d", status, resp.Code, errs.CodeInvalidParam)
	}
	if fake.createCalls != 0 {
		t.Errorf("svc.Create called %d times, want 0", fake.createCalls)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	fake := &fakeProductService{
		getErr: errs.New(errs.CodeProductNotFound, "product not found", http.StatusNotFound),
	}
	r := newTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodGet, "/api/v1/products/999", "")
	if status != http.StatusNotFound || resp.Code != errs.CodeProductNotFound {
		t.Fatalf("status=%d code=%d, want 404/%d", status, resp.Code, errs.CodeProductNotFound)
	}
}

func TestGetByIDInvalidParam(t *testing.T) {
	r := newTestRouter(&fakeProductService{})

	status, resp := doRequest(t, r, http.MethodGet, "/api/v1/products/abc", "")
	if status != http.StatusBadRequest || resp.Code != errs.CodeInvalidParam {
		t.Fatalf("status=%d code=%d, want 400/%d", status, resp.Code, errs.CodeInvalidParam)
	}
}

func TestHealthz(t *testing.T) {
	r := newTestRouter(&fakeProductService{})

	status, resp := doRequest(t, r, http.MethodGet, "/healthz", "")
	if status != http.StatusOK || resp.Code != errs.CodeOK {
		t.Fatalf("status=%d code=%d, want 200/0", status, resp.Code)
	}
}
