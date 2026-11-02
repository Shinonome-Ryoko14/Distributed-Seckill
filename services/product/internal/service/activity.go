// Package service 实现 product 领域活动业务逻辑。
package service

import (
	"context"
	"net/http"
	"time"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/repository"
)

// CreateActivityInput 是创建秒杀活动的业务入参。
type CreateActivityInput struct {
	ProductID    uint64
	SeckillStock uint32
	SeckillPrice uint32
	StartTime    time.Time
	EndTime      time.Time
	Status       model.ActivityStatus
}

// ActivityService 提供秒杀活动主数据业务。
type ActivityService struct {
	products   repository.ProductRepository
	activities repository.ActivityRepository
	now        func() time.Time
}

// NewActivityService 构造 ActivityService，注入商品与活动数据访问依赖和时钟依赖。
func NewActivityService(products repository.ProductRepository, activities repository.ActivityRepository, now func() time.Time) *ActivityService {
	return &ActivityService{products: products, activities: activities, now: now}
}

// Create 校验入参并创建秒杀活动，返回新活动 ID。
func (s *ActivityService) Create(ctx context.Context, in CreateActivityInput) (uint64, error) {
	p, err := s.products.GetByID(ctx, in.ProductID)
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, errs.New(errs.CodeProductNotFound, "product not found", http.StatusNotFound)
	}

	now := s.now()
	err = validateCreateActivity(in, p, now)
	if err != nil {
		return 0, err
	}
	a := &model.Activity{
		ProductID:    in.ProductID,
		SeckillStock: in.SeckillStock,
		SeckillPrice: in.SeckillPrice,
		StartTime:    in.StartTime,
		EndTime:      in.EndTime,
		Status:       in.Status,
	}
	id, err := s.activities.Create(ctx, a)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func validateCreateActivity(in CreateActivityInput, p *model.Product, now time.Time) error {
	stock := in.SeckillStock
	price := in.SeckillPrice
	if p.Status != model.StatusOnSale {
		return errs.New(errs.CodeActivityInvalid, "product is not on sale", http.StatusBadRequest)
	}
	if stock == 0 {
		return errs.New(errs.CodeActivityInvalid, "invalid seckill stock", http.StatusBadRequest)
	}
	if price == 0 {
		return errs.New(errs.CodeActivityInvalid, "invalid seckill price", http.StatusBadRequest)
	}

	if price >= p.Price {
		return errs.New(errs.CodeActivityInvalid, "seckill price must be lower than product price",
			http.StatusBadRequest)
	}
	if in.StartTime.IsZero() || in.EndTime.IsZero() || !in.StartTime.Before(in.EndTime) {
		return errs.New(errs.CodeActivityInvalid, "invalid activity time range", http.StatusBadRequest)
	}
	switch in.Status {
	case model.ActivityStatusPending, model.ActivityStatusActive, model.ActivityStatusCompleted:
	default:
		return errs.New(errs.CodeActivityInvalid, "invalid activity status", http.StatusBadRequest)
	}
	return validateActivityStatusTime(in, now)
}

func validateActivityStatusTime(in CreateActivityInput, now time.Time) error {
	switch in.Status {
	case model.ActivityStatusPending:
		if !now.Before(in.StartTime) {
			return errs.New(errs.CodeActivityInvalid, "activity status is not pending", http.StatusBadRequest)
		}
	case model.ActivityStatusActive:
		if now.Before(in.StartTime) || !now.Before(in.EndTime) {
			return errs.New(errs.CodeActivityInvalid, "activity status is not active", http.StatusBadRequest)
		}
	case model.ActivityStatusCompleted:
		if now.Before(in.EndTime) {
			return errs.New(errs.CodeActivityInvalid, "activity status is not completed", http.StatusBadRequest)
		}
	default:
		return errs.New(errs.CodeActivityInvalid, "invalid activity status", http.StatusBadRequest)
	}
	return nil
}

// GetByID 按 ID 查询活动；不存在返回 CodeActivityNotFound(404)。
func (s *ActivityService) GetByID(ctx context.Context, id uint64) (*model.Activity, error) {
	a, err := s.activities.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errs.New(errs.CodeActivityNotFound, "activity not found", http.StatusNotFound)
	}
	return a, nil
}
