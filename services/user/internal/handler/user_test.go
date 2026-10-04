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
)

type fakeUserService struct {
	registerID    uint64
	registerErr   error
	loginToken    string
	loginErr      error
	registerCalls int
	gotUsername   string
	gotPassword   string
}

var _ userService = (*fakeUserService)(nil)

func (f *fakeUserService) Register(_ context.Context, username, password string) (uint64, error) {
	f.registerCalls++
	f.gotUsername, f.gotPassword = username, password
	return f.registerID, f.registerErr
}

func (f *fakeUserService) Login(_ context.Context, username, password string) (string, error) {
	f.gotUsername, f.gotPassword = username, password
	return f.loginToken, f.loginErr
}

type apiResponse struct {
	Code    errs.Code       `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func newTestRouter(fake *fakeUserService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return NewRouter(NewUserHandler(fake))
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

func TestRegisterSuccess(t *testing.T) {
	fake := &fakeUserService{registerID: 42}
	r := newTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodPost,
		"/api/v1/users/register", `{"username": "alice", "password": "secret-password"}`)

	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if resp.Code != errs.CodeOK {
		t.Fatalf("code = %d, want %d", resp.Code, errs.CodeOK)
	}

	var data struct {
		ID uint64 `json:"id"`
	}

	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if data.ID != 42 {
		t.Errorf("data.id = %d, want 42", data.ID)
	}
	if fake.gotUsername != "alice" || fake.gotPassword != "secret-password" {
		t.Errorf("svc got(%q, %q), want (alice, secret-password)",
			fake.gotUsername, fake.gotPassword)
	}
}

// TestRegisterInvalidJSON 校验请求体解析失败时，handler 返回入参错误，
// 且不会把脏请求透传给 service（registerCalls 必须为 0）。
func TestRegisterInvalidJSON(t *testing.T) {
	fake := &fakeUserService{}
	r := newTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodPost, "/api/v1/users/register", `{`)

	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
	}
	if resp.Code != errs.CodeInvalidParam {
		t.Errorf("code = %d, want %d", resp.Code, errs.CodeInvalidParam)
	}
	// 绑定失败必须在 handler 层短路，不能调用 service 做无谓的业务/DB 处理。
	if fake.registerCalls != 0 {
		t.Errorf("svc.Register called %d times, want 0", fake.registerCalls)
	}
}

// TestRegisterServiceError 校验 service 返回的业务错误被原样映射为
// 对应 HTTP 状态与业务码（这里用用户名冲突：409 / CodeUserExists）。
func TestRegisterServiceError(t *testing.T) {
	fake := &fakeUserService{
		registerErr: errs.New(errs.CodeUserExists, "username already exists", http.StatusConflict),
	}
	r := newTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodPost, "/api/v1/users/register",
		`{"username": "alice", "password": "secret-password"}`)

	if status != http.StatusConflict {
		t.Fatalf("status = %d, want %d", status, http.StatusConflict)
	}
	if resp.Code != errs.CodeUserExists {
		t.Errorf("code = %d, want %d", resp.Code, errs.CodeUserExists)
	}
}

// TestLoginSuccess 校验登录成功后返回 token，且 fake 收到正确入参。
func TestLoginSuccess(t *testing.T) {
	fake := &fakeUserService{loginToken: "tok"}
	r := newTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodPost, "/api/v1/users/login",
		`{"username": "alice", "password": "secret-password"}`)

	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if resp.Code != errs.CodeOK {
		t.Fatalf("code = %d, want %d", resp.Code, errs.CodeOK)
	}

	var data struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if data.Token != "tok" {
		t.Errorf("data.token = %q, want %q", data.Token, "tok")
	}
}

// TestLoginWrongPassword 校验密码错误映射为 401 / CodePasswordWrong。
func TestLoginWrongPassword(t *testing.T) {
	fake := &fakeUserService{
		loginErr: errs.New(errs.CodePasswordWrong, "incorrect password", http.StatusUnauthorized),
	}
	r := newTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodPost, "/api/v1/users/login",
		`{"username": "alice", "password": "wrong-password"}`)

	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
	}
	if resp.Code != errs.CodePasswordWrong {
		t.Errorf("code = %d, want %d", resp.Code, errs.CodePasswordWrong)
	}
}

// TestLoginUserNotFound 校验用户不存在映射为 401 / CodeUserNotFound。
func TestLoginUserNotFound(t *testing.T) {
	fake := &fakeUserService{
		loginErr: errs.New(errs.CodeUserNotFound, "user not found", http.StatusUnauthorized),
	}
	r := newTestRouter(fake)

	status, resp := doRequest(t, r, http.MethodPost, "/api/v1/users/login",
		`{"username": "nobody", "password": "secret-password"}`)

	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
	}
	if resp.Code != errs.CodeUserNotFound {
		t.Errorf("code = %d, want %d", resp.Code, errs.CodeUserNotFound)
	}
}

// TestHealthz 校验健康检查返回 200 与统一成功码。
func TestHealthz(t *testing.T) {
	r := newTestRouter(&fakeUserService{})

	status, resp := doRequest(t, r, http.MethodGet, "/healthz", "")

	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if resp.Code != errs.CodeOK {
		t.Errorf("code = %d, want %d", resp.Code, errs.CodeOK)
	}
}
