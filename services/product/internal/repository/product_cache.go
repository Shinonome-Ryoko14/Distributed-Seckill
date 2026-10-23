package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
)

// ProductCache 抽象商品详情缓存，便于替换实现与单测。
type ProductCache interface {
	Get(ctx context.Context, id uint64) (*model.Product, bool, error)

	Set(ctx context.Context, id uint64, p *model.Product, ttl time.Duration) error

	Delete(ctx context.Context, id uint64) error
}

// CachedProductRepository 用 Cache Aside 装饰 ProductRepository，只缓存单商品读取。
type CachedProductRepository struct {
	repo    ProductRepository
	cache   ProductCache
	ttl     time.Duration
	nullTTL time.Duration
}

// NewCachedProductRepository 构造带缓存的商品仓库。
func NewCachedProductRepository(
	repo ProductRepository,
	cache ProductCache,
	ttl time.Duration,
	nullTTL time.Duration,
) *CachedProductRepository {
	return &CachedProductRepository{
		repo:    repo,
		cache:   cache,
		ttl:     ttl,
		nullTTL: nullTTL,
	}
}

// Create 透传写入，并清掉新 id 上可能存在的负缓存。
func (r *CachedProductRepository) Create(ctx context.Context, p *model.Product) (uint64, error) {
	id, err := r.repo.Create(ctx, p)
	if err != nil {
		return 0, err
	}

	err = r.cache.Delete(ctx, id)
	if err != nil {
		slog.Warn("delete product cache after create", "id", id, "error", err)
	}
	return id, nil
}

// GetByID 先读缓存，未命中回源并按结果回填。
func (r *CachedProductRepository) GetByID(ctx context.Context, id uint64) (*model.Product, error) {
	p, hit, err := r.cache.Get(ctx, id)
	if err != nil {
		slog.Warn("get product cache", "id", id, "error", err)
	} else /*如果缓存命中*/ if hit {
		return p, nil
	}

	// 如果缓存未命中
	p, err = r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 如果该物品不存在
	if p == nil {
		if err := r.cache.Set(ctx, id, nil, r.nullTTL); err != nil {
			slog.Warn("set product null cache", "id", id, "error", err)
		}
		return nil, nil
	}

	// 如果该物品存在
	if err := r.cache.Set(ctx, id, p, r.ttl); err != nil {
		slog.Warn("set product cache", "id", id, "error", err)
	}
	return p, nil
}

// List 透传，不做列表缓存。
func (r *CachedProductRepository) List(ctx context.Context, limit, offset int) ([]*model.Product, error) {
	return r.repo.List(ctx, limit, offset)
}

// Count 透传。
func (r *CachedProductRepository) Count(ctx context.Context) (int64, error) {
	return r.repo.Count(ctx)
}
