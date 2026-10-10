// Package handler 提供 product 服务的 HTTP 接入层。
package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/response"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/service"
)

type productService interface {
	Create(ctx context.Context, in service.CreateProductInput) (uint64, error)
	GetByID(ctx context.Context, id uint64) (*model.Product, error)
	List(ctx context.Context, page, size int) (*service.ProductPage, error)
}

// ProductHandler 处理商品相关 HTTP 请求。
type ProductHandler struct {
	svc productService
}

// NewProductHandler 构造 ProductHandler。
func NewProductHandler(svc productService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

type createProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       uint32 `json:"price"`
	Status      int8   `json:"status"`
}

// Create 处理 POST /api/v1/products。
func (h *ProductHandler) Create(c *gin.Context) {
	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c.Writer, errs.New(errs.CodeInvalidParam, "invalid param", http.StatusBadRequest))
		return
	}

	id, err := h.svc.Create(c.Request.Context(), service.CreateProductInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Status:      model.Status(req.Status),
	})
	if err != nil {
		response.Fail(c.Writer, err)
		return
	}
	response.OK(c.Writer, gin.H{"id": id})
}

// GetByID 处理 GET /api/v1/products/{id}。
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c.Writer, errs.New(errs.CodeInvalidParam, "invalid id", http.StatusBadRequest))
		return
	}

	p, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c.Writer, err)
		return
	}
	response.OK(c.Writer, p)
}

// List 处理 GET /api/v1/products?page=1&size=20。
func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	p, err := h.svc.List(c.Request.Context(), page, size)
	if err != nil {
		response.Fail(c.Writer, err)
		return
	}

	response.OK(c.Writer, gin.H{
		"items": p.Items,
		"total": p.Total,
		"page":  p.Page,
		"size":  p.Size,
	})
}
