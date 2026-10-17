package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/response"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/service"
)

type activityService interface {
	Create(ctx context.Context, in service.CreateActivityInput) (uint64, error)
	GetByID(ctx context.Context, id uint64) (*model.Activity, error)
}

// ActivityHandler 处理秒杀活动相关 HTTP 请求。
type ActivityHandler struct {
	svc activityService
}

// NewActivityHandler 构造 ActivityHandler。
func NewActivityHandler(svc activityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

type createActivityRequest struct {
	ProductID    uint64               `json:"product_id"`
	SeckillStock uint32               `json:"seckill_stock"`
	SeckillPrice uint32               `json:"seckill_price"`
	StartTime    string               `json:"start_time"`
	EndTime      string               `json:"end_time"`
	Status       model.ActivityStatus `json:"status"`
}

// Create 处理 POST /api/v1/activities。
func (h *ActivityHandler) Create(c *gin.Context) {
	var req createActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c.Writer, errs.New(errs.CodeInvalidParam, "invalid param", http.StatusBadRequest))
		return
	}

	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		response.Fail(c.Writer, errs.New(errs.CodeInvalidParam, "invalid start_time", http.StatusBadRequest))
		return
	}

	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		response.Fail(c.Writer, errs.New(errs.CodeInvalidParam, "invalid end_time", http.StatusBadRequest))
		return
	}

	id, err := h.svc.Create(c.Request.Context(), service.CreateActivityInput{
		ProductID:    req.ProductID,
		SeckillStock: req.SeckillStock,
		SeckillPrice: req.SeckillPrice,
		StartTime:    start,
		EndTime:      end,
		Status:       req.Status,
	})
	if err != nil {
		response.Fail(c.Writer, err)
		return
	}
	response.OK(c.Writer, gin.H{"id": id})
}

// GetByID 处理 GET /api/v1/activities/{id}。
func (h *ActivityHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c.Writer, errs.New(errs.CodeInvalidParam, "invalid id", http.StatusBadRequest))
		return
	}
	a, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c.Writer, err)
		return
	}

	response.OK(c.Writer, a)
}
