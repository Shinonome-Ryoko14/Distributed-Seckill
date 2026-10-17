package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/response"
)

// NewRouter 装配 product 服务的路由与中间件。
func NewRouter(ph *ProductHandler, ah *ActivityHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		response.OK(c.Writer, nil)
	})

	products := r.Group("/api/v1/products")
	{
		products.POST("", ph.Create)
		products.GET("", ph.List)
		products.GET("/:id", ph.GetByID)
	}

	activities := r.Group("/api/v1/activities")
	{
		activities.POST("", ah.Create)
		activities.GET("/:id", ah.GetByID)
	}
	return r
}
