package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/service"
)

type fakeActivityService struct {
	createID    uint64
	createErr   error
	getResult   *model.Activity
	getErr      error
	createCalls int
	lastInput   service.CreateActivityInput
}

var _ activityService = (*fakeActivityService)(nil)

func (f *fakeActivityService) Create(_ context.Context, in service.CreateActivityInput) (uint64, error) {
	f.createCalls++
	f.lastInput = in
	return f.createID, f.createErr
}

func (f *fakeActivityService) GetByID(_ context.Context, _ uint64) (*model.Activity, error) {
	return f.getResult, f.getErr
}

func newActivityTestRouter(fake *fakeActivityService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return NewRouter(NewProductHandler(&fakeProductService{}), NewActivityHandler(fake))
}

const validCreateActivityBody = `{"product_id":1,"seckill_stock":10,"seckill_price":500,"start_time":"2026-10-17T10:00:00Z","end_time":"2026-10-17T12:00:00Z","status":1}`

func TestCreateActivitySuccess(t *testing.T) {
	fake := &fakeActivityService{createID: 11}
	r := newActivityTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodPost, "/api/v1/activities", validCreateActivityBody)
	if status != http.StatusOK || resp.Code != errs.CodeOK {
		t.Fatalf("status=%d code=%d, want 200/0", status, resp.Code)
	}

	var data struct {
		ID uint64 `json:"id"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if data.ID != 11 {
		t.Errorf("data.id = %d, want 11", data.ID)
	}
	if fake.createCalls != 1 {
		t.Errorf("svc.Create calls = %d, want 1", fake.createCalls)
	}
	if fake.lastInput.StartTime.IsZero() || fake.lastInput.EndTime.IsZero() {
		t.Errorf("times not parsed: %+v", fake.lastInput)
	}
}

func TestCreateActivityInvalidJSON(t *testing.T) {
	fake := &fakeActivityService{}
	r := newActivityTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodPost, "/api/v1/activities", `{`)
	if status != http.StatusBadRequest || resp.Code != errs.CodeInvalidParam {
		t.Fatalf("status=%d code=%d, want 400/%d", status, resp.Code, errs.CodeInvalidParam)
	}
	if fake.createCalls != 0 {
		t.Errorf("svc.Create called %d times, want 0", fake.createCalls)
	}
}

func TestCreateActivityInvalidTime(t *testing.T) {
	fake := &fakeActivityService{}
	r := newActivityTestRouter(fake)

	body := `{"product_id":1,"seckill_stock":10,"seckill_price":500,"start_time":"not-a-time","end_time":"2026-10-17T12:00:00Z","status":1}`
	status, resp := doRequest(t, r, http.MethodPost, "/api/v1/activities", body)
	if status != http.StatusBadRequest || resp.Code != errs.CodeInvalidParam {
		t.Fatalf("status=%d code=%d, want 400/%d", status, resp.Code, errs.CodeInvalidParam)
	}
	if fake.createCalls != 0 {
		t.Errorf("svc.Create called %d times, want 0", fake.createCalls)
	}
}

func TestCreateActivityServiceError(t *testing.T) {
	fake := &fakeActivityService{
		createErr: errs.New(errs.CodeProductNotFound, "product not found", http.StatusNotFound),
	}
	r := newActivityTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodPost, "/api/v1/activities", validCreateActivityBody)
	if status != http.StatusNotFound || resp.Code != errs.CodeProductNotFound {
		t.Fatalf("status=%d code=%d, want 404/%d", status, resp.Code, errs.CodeProductNotFound)
	}
}

func TestGetActivitySuccess(t *testing.T) {
	fake := &fakeActivityService{getResult: &model.Activity{ID: 9, ProductID: 1}}
	r := newActivityTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodGet, "/api/v1/activities/9", "")
	if status != http.StatusOK || resp.Code != errs.CodeOK {
		t.Fatalf("status=%d code=%d, want 200/0", status, resp.Code)
	}

	var data model.Activity
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if data.ID != 9 {
		t.Errorf("data.id = %d, want 9", data.ID)
	}
}

func TestGetActivityNotFound(t *testing.T) {
	fake := &fakeActivityService{
		getErr: errs.New(errs.CodeActivityNotFound, "activity not found", http.StatusNotFound),
	}
	r := newActivityTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodGet, "/api/v1/activities/999", "")
	if status != http.StatusNotFound || resp.Code != errs.CodeActivityNotFound {
		t.Fatalf("status=%d code=%d, want 404/%d", status, resp.Code, errs.CodeActivityNotFound)
	}
}

func TestGetActivityInvalidID(t *testing.T) {
	r := newActivityTestRouter(&fakeActivityService{})

	status, resp := doRequest(t, r, http.MethodGet, "/api/v1/activities/abc", "")
	if status != http.StatusBadRequest || resp.Code != errs.CodeInvalidParam {
		t.Fatalf("status=%d code=%d, want 400/%d", status, resp.Code, errs.CodeInvalidParam)
	}
}
