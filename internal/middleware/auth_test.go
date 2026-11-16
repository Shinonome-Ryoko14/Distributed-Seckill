package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/ctxkey"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/jwt"
)

func newPrivateRouter(manager *jwt.Manager, nextCalled *bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Auth(manager))
	r.GET("/private", func(c *gin.Context) {
		*nextCalled = true
		c.Status(http.StatusOK)
	})
	return r
}

func TestAuthRejectsBadAuthorization(t *testing.T) {
	manager := jwt.NewManager("secret", time.Hour, "issuer")

	other := jwt.NewManager("other-secret", time.Hour, "issuer")
	otherToken, err := other.Generate(1)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	cases := []struct {
		name   string
		header string
	}{
		{"missing", ""},
		{"wrong scheme", "Token abc"},
		{"empty token", "Bearer "},
		{"malformed token", "Bearer not-a-jwt"},
		{"wrong secret", "Bearer " + otherToken},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false
			r := newPrivateRouter(manager, &nextCalled)

			req := httptest.NewRequestWithContext(
				context.Background(), http.MethodGet, "/private", nil,
			)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", w.Code)
			}

			var body struct {
				Code errs.Code `json:"code"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if body.Code != errs.CodeUnauthorized {
				t.Errorf("code = %d, want %d", body.Code, errs.CodeUnauthorized)
			}
			if got := w.Header().Get("WWW-Authenticate"); got != "Bearer" {
				t.Errorf("WWW-Authenticate = %q, want Bearer", got)
			}
			if nextCalled {
				t.Error("next handler must not run after auth failure")
			}
		})
	}
}

func TestAuthInjectsUserID(t *testing.T) {
	manager := jwt.NewManager("secret", time.Hour, "issuer")
	token, err := manager.Generate(42)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	var gotUserID uint64
	var gotOK bool

	r := gin.New()
	r.Use(Auth(manager))
	r.GET("/private", func(c *gin.Context) {
		gotUserID, gotOK = ctxkey.UserID(c.Request.Context())
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet, "/private", nil,
	)
	// 小写 bearer 也能通过，验证 scheme 大小写不敏感。
	req.Header.Set("Authorization", "bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !gotOK || gotUserID != 42 {
		t.Errorf("ctx user id = (%d, %v), want (42, true)", gotUserID, gotOK)
	}
}
