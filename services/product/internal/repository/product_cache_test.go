package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
)

type fakeProductCache struct {
	values    map[uint64]*model.Product
	hits      map[uint64]bool
	setTTLs   map[uint64]time.Duration
	deleted   []uint64
	getErr    error
	setErr    error
	deleteErr error
}

func newFakeProductCache() *fakeProductCache {
	return &fakeProductCache{
		values:  make(map[uint64]*model.Product),
		hits:    make(map[uint64]bool),
		setTTLs: make(map[uint64]time.Duration),
	}
}

func (f *fakeProductCache) Get(_ context.Context, id uint64) (*model.Product, bool, error) {
	if f.getErr != nil {
		return nil, false, f.getErr
	}
	if !f.hits[id] {
		return nil, false, nil
	}
	return f.values[id], true, nil
}

func (f *fakeProductCache) Set(_ context.Context, id uint64, p *model.Product, ttl time.Duration) error {
	if f.setErr != nil {
		return f.setErr
	}
	f.values[id] = p
	f.hits[id] = true
	f.setTTLs[id] = ttl
	return nil
}

func (f *fakeProductCache) Delete(_ context.Context, id uint64) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.values, id)
	delete(f.hits, id)
	f.deleted = append(f.deleted, id)
	return nil
}

type stubProductRepo struct {
	product   *model.Product
	getCalls  int
	createdID uint64
	createErr error
}

func (s *stubProductRepo) Create(_ context.Context, _ *model.Product) (uint64, error) {
	return s.createdID, s.createErr
}

func (s *stubProductRepo) GetByID(_ context.Context, _ uint64) (*model.Product, error) {
	s.getCalls++
	return s.product, nil
}

func (s *stubProductRepo) List(_ context.Context, _, _ int) ([]*model.Product, error) {
	return nil, nil
}

func (s *stubProductRepo) Count(_ context.Context) (int64, error) { return 0, nil }

func TestCachedGetByIDHit(t *testing.T) {
	t.Skip("TODO(user): 缓存命中时不查库，返回缓存对象")
}

func TestCachedGetByIDMissFillsCache(t *testing.T) {
	t.Skip("TODO(user): 未命中回源，按 ttl 回填")
}

func TestCachedGetByIDNotFoundCachesNull(t *testing.T) {
	t.Skip("TODO(user): 商品不存在时写 nil 值，TTL 为 nullTTL")
}

func TestCachedGetByIDCacheErrorFallsBack(t *testing.T) {
	t.Skip("TODO(user): cache.Get 返回错误时仍查库并返回成功")
}

func TestCachedCreateDeletesNegativeCache(t *testing.T) {
	t.Skip("TODO(user): Create 成功后 deleted 包含新 id")
}
