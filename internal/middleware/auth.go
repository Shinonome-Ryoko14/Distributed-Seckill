// Package middleware 提供跨服务复用的 Gin 中间件。
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/ctxkey"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/jwt"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/response"
)

const bearerPrefix = "Bearer "

// Auth 校验 Authorization: Bearer <token>，成功后把 user_id 注入 request context。
func Auth(manager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			unauthorized(c)
			return
		}

		claims, err := manager.Parse(token)
		if err != nil {
			unauthorized(c)
			return
		}

		ctx := ctxkey.WithUserID(c.Request.Context(), claims.UserID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	c.Header("WWW-Authenticate", "Bearer")
	response.Fail(c.Writer, errs.New(errs.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
	c.Abort()
}

func bearerToken(header string) (string, bool) {
	if len(header) < len(bearerPrefix) ||
		!strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(bearerPrefix):])
	return token, token != ""
}
