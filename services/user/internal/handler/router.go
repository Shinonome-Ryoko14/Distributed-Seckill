package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/response"
)

// NewRouter 装配 user 服务的路由与中间件。
func NewRouter(h *UserHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		response.OK(c.Writer, nil)
	})

	users := r.Group("/api/v1/users")
	{
		users.POST("/register", h.Register)
		users.POST("/login", h.Login)
	}
	return r
}
