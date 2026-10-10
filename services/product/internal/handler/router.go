package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/response"
)

// NewRouter 装配 product 服务的路由与中间件。
func NewRouter(h *ProductHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		response.OK(c.Writer, nil)
	})

	products := r.Group("/api/v1/products")
	{
		products.POST("", h.Create)
		products.GET("", h.List)
		products.GET("/:id", h.GetByID)
	}
	return r
}
