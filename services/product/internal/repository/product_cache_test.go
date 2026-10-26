package repository

import (
	"context"
	"errors"
	"slices"
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
	getErr    error
}

func (s *stubProductRepo) Create(_ context.Context, _ *model.Product) (uint64, error) {
	return s.createdID, s.createErr
}

func (s *stubProductRepo) GetByID(_ context.Context, _ uint64) (*model.Product, error) {
	s.getCalls++
	return s.product, s.getErr
}

func (s *stubProductRepo) List(_ context.Context, _, _ int) ([]*model.Product, error) {
	return nil, nil
}

func (s *stubProductRepo) Count(_ context.Context) (int64, error) { return 0, nil }

func TestCachedGetByIDHit(t *testing.T) {
	const id = uint64(1)

	cached := &model.Product{ID: id, Name: "from-cache"}
	repo := &stubProductRepo{
		product: &model.Product{ID: id, Name: "from-db"},
	}

	cache := newFakeProductCache()
	cache.values[id] = cached
	cache.hits[id] = true

	const ttl = 5 * time.Minute
	const nullTTL = time.Minute
	svc := NewCachedProductRepository(repo, cache, ttl, nullTTL)

	got, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v, want nil", err)
	}
	if got != cached {
		t.Errorf("GetByID() = %+v, want cached product %+v", got, cached)
	}
	if repo.getCalls != 0 {
		t.Errorf("repo.GetByID calls = %d, want 0", repo.getCalls)
	}
}

func TestCachedGetByIDMissFillsCache(t *testing.T) {
	const id = uint64(2)

	dbProduct := &model.Product{ID: id, Name: "from-db"}
	repo := &stubProductRepo{product: dbProduct}
	cache := newFakeProductCache()

	const ttl = 5 * time.Minute
	const nullTTL = time.Minute
	svc := NewCachedProductRepository(repo, cache, ttl, nullTTL)

	got, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v, want nil", err)
	}
	if got != dbProduct {
		t.Errorf("GetByID() = %+v, want db product %+v", got, dbProduct)
	}
	if repo.getCalls != 1 {
		t.Errorf("repo.GetByID calls = %d, want 1", repo.getCalls)
	}
	if !cache.hits[id] {
		t.Errorf("cache.hits[%d] = false, want true (cache not filled)", id)
	}
	if cache.values[id] != dbProduct {
		t.Errorf("cache.values[%d] = %+v, want db product %+v", id, cache.values[id], dbProduct)
	}
	if cache.setTTLs[id] != ttl {
		t.Errorf("cache TTL = %s, want %s", cache.setTTLs[id], ttl)
	}
}

func TestCachedGetByIDNotFoundCachesNull(t *testing.T) {
	const id = uint64(3)

	repo := &stubProductRepo{product: nil}
	cache := newFakeProductCache()

	const ttl = 5 * time.Minute
	const nullTTL = time.Minute

	svc := NewCachedProductRepository(repo, cache, ttl, nullTTL)

	got, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v, want nil", err)
	}
	if got != nil {
		t.Errorf("GetByID() = %+v, want nil", got)
	}
	if repo.getCalls != 1 {
		t.Errorf("repo.GetByID calls = %d, want 1", repo.getCalls)
	}
	if !cache.hits[id] {
		t.Errorf("negative cache not written: hits[%d] = false", id)
	}
	if cache.values[id] != nil {
		t.Errorf("cache.values[%d] = %+v, want nil", id, cache.values[id])
	}
	if cache.setTTLs[id] != nullTTL {
		t.Errorf("null cache TTL = %s, want %s", cache.setTTLs[id], nullTTL)
	}
}

func TestCachedGetByIDCacheErrorFallsBack(t *testing.T) {
	const id = uint64(4)

	dbProduct := &model.Product{ID: id, Name: "from-db"}
	repo := &stubProductRepo{product: dbProduct}
	cache := newFakeProductCache()
	cache.getErr = errors.New("redis down")

	const ttl = 5 * time.Minute
	const nullTTL = time.Minute
	svc := NewCachedProductRepository(repo, cache, ttl, nullTTL)

	got, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v, want nil", err)
	}
	if got != dbProduct {
		t.Errorf("GetByID() = %+v, want db product %+v", got, dbProduct)
	}
	if repo.getCalls != 1 {
		t.Errorf("repo.GetByID calls = %d, want 1", repo.getCalls)
	}
	if cache.setTTLs[id] != ttl {
		t.Errorf("cache TTL = %s, want %s (should try to refill cache)", cache.setTTLs[id], ttl)
	}
}

func TestCachedCreateDeletesNegativeCache(t *testing.T) {
	const newID = uint64(42)

	repo := &stubProductRepo{createdID: newID}
	cache := newFakeProductCache()

	const ttl = 5 * time.Minute
	const nullTTL = time.Minute
	svc := NewCachedProductRepository(repo, cache, ttl, nullTTL)
	id, err := svc.Create(context.Background(), &model.Product{Name: "new"})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if id != newID {
		t.Errorf("Create() id = %d, want %d", id, newID)
	}
	if !slices.Contains(cache.deleted, newID) {
		t.Errorf("cache.Delete(%d) not called, deleted = %v", newID, cache.deleted)
	}
}

func TestCachedGetByIDDBErrorDoesNotCache(t *testing.T) {
	const id = uint64(5)

	dbErr := errors.New("db down")
	repo := &stubProductRepo{getErr: dbErr}
	cache := newFakeProductCache()

	const ttl = 5 * time.Minute
	const nullTTL = time.Minute
	svc := NewCachedProductRepository(repo, cache, ttl, nullTTL)

	got, err := svc.GetByID(context.Background(), id)
	if !errors.Is(err, dbErr) {
		t.Fatalf("GetByID() error = %v, want %v", err, dbErr)
	}
	if got != nil {
		t.Errorf("GetByID() = %+v, want nil", got)
	}
	if cache.hits[id] {
		t.Errorf("negative cache should not be written on DB error")
	}
}
