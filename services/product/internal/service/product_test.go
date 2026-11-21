package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/repository"
)

type fakeProductRepo struct {
	byID        map[uint64]*model.Product
	nextID      uint64
	createCalls int
	listErr     error
	lastLimit   int
	lastOffset  int
}

var _ repository.ProductRepository = (*fakeProductRepo)(nil)

func newFakeProductRepo() *fakeProductRepo {
	return &fakeProductRepo{byID: map[uint64]*model.Product{}}
}

func (f *fakeProductRepo) Create(_ context.Context, p *model.Product) (uint64, error) {
	f.createCalls++
	f.nextID++
	stored := *p
	stored.ID = f.nextID
	f.byID[stored.ID] = &stored
	return stored.ID, nil
}

func (f *fakeProductRepo) GetByID(_ context.Context, id uint64) (*model.Product, error) {
	return f.byID[id], nil
}

func (f *fakeProductRepo) List(_ context.Context, limit, offset int) ([]*model.Product, error) {
	f.lastLimit, f.lastOffset = limit, offset
	if f.listErr != nil {
		return nil, f.listErr
	}
	items := make([]*model.Product, 0, len(f.byID))
	for _, p := range f.byID {
		items = append(items, p)
	}
	return items, nil
}

func (f *fakeProductRepo) Count(_ context.Context) (int64, error) {
	return int64(len(f.byID)), nil
}

func appErrCode(t *testing.T, err error) errs.Code {
	t.Helper()
	var appErr *errs.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %v, want *errs.AppError", err)
	}
	return appErr.Code
}

func TestCreateSuccess(t *testing.T) {
	repo := newFakeProductRepo()
	svc := NewProductService(repo)

	id, err := svc.Create(context.Background(), CreateProductInput{
		Name: "iPhone", Description: "demo", Price: 100, Status: model.StatusOnSale,
	})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if id == 0 {
		t.Fatalf("Create() id = 0, want non-zero")
	}
	if repo.createCalls != 1 {
		t.Errorf("repo.Create calls = %d, want 1", repo.createCalls)
	}
}

func TestCreateInvalidParam(t *testing.T) {
	cases := []struct {
		name string
		in   CreateProductInput
	}{
		{"empty name", CreateProductInput{Name: " ", Price: 1, Status: model.StatusOnSale}},
		{"zero price", CreateProductInput{Name: "ok", Price: 0, Status: model.StatusOnSale}},
		{"bad status", CreateProductInput{Name: "ok", Price: 1, Status: model.Status(9)}},
		{"name too long", CreateProductInput{Name: strings.Repeat("a", 65), Price: 1, Status: model.StatusOnSale}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeProductRepo()
			svc := NewProductService(repo)

			_, err := svc.Create(context.Background(), tc.in)
			if got := appErrCode(t, err); got != errs.CodeProductInvalid {
				t.Errorf("Create() code = %d, want %d", got, errs.CodeProductInvalid)
			}
			if repo.createCalls != 0 {
				t.Errorf("repo.createCalls = %d, want 0", repo.createCalls)
			}
		})
	}
}

func TestGetByIDNotFound(t *testing.T) {
	svc := NewProductService(newFakeProductRepo())

	_, err := svc.GetByID(context.Background(), 999)
	if got := appErrCode(t, err); got != errs.CodeProductNotFound {
		t.Errorf("GetByID() code = %d, want %d", got, errs.CodeProductNotFound)
	}
}

func TestGetByIDSuccess(t *testing.T) {
	repo := newFakeProductRepo()
	svc := NewProductService(repo)
	id, _ := svc.Create(context.Background(), CreateProductInput{
		Name: "p", Price: 1, Status: model.StatusOnSale,
	})

	p, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v, want nil", err)
	}
	if p == nil || p.ID != id {
		t.Fatalf("GetByID() = %+v, want id %d", p, id)
	}
}

func TestListNormalizePage(t *testing.T) {
	repo := newFakeProductRepo()
	svc := NewProductService(repo)
	for i := 0; i < 3; i++ {
		if _, err := repo.Create(context.Background(), &model.Product{Name: "p"}); err != nil {
			t.Fatalf("seed error = %v", err)
		}
	}

	page, err := svc.List(context.Background(), 0, 1000)
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if page.Page != 1 {
		t.Errorf("Page = %d, want 1", page.Page)
	}
	if page.Size != maxPageSize {
		t.Errorf("Size = %d, want %d", page.Size, maxPageSize)
	}
	if page.Total != 3 {
		t.Errorf("Total = %d, want 3", page.Total)
	}
}

func TestNormalizePageClampsMax(t *testing.T) {
	page, size := normalizePage(math.MaxInt, maxPageSize)
	if page != maxPage {
		t.Errorf("page = %d, want %d", page, maxPage)
	}
	if size != maxPageSize {
		t.Errorf("size = %d, want %d", size, maxPageSize)
	}
}

func TestListHugePageOffsetDoesNotOverflow(t *testing.T) {
	repo := newFakeProductRepo()
	svc := NewProductService(repo)

	got, err := svc.List(context.Background(), math.MaxInt, maxPageSize)
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if got.Page != maxPage {
		t.Errorf("Page = %d, want %d", got.Page, maxPage)
	}

	wantOffset := (maxPage - 1) * maxPageSize
	if repo.lastLimit != maxPageSize || repo.lastOffset != wantOffset {
		t.Errorf("repo.List args = (%d,%d), want (%d,%d)",
			repo.lastLimit, repo.lastOffset, maxPageSize, wantOffset)
	}
	if repo.lastOffset < 0 {
		t.Errorf("offset = %d, must not be negative", repo.lastOffset)
	}
}
