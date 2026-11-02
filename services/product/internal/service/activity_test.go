package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/repository"
)

type fakeActivityRepo struct {
	byID        map[uint64]*model.Activity
	nextID      uint64
	createCalls int
	createErr   error
	getErr      error
}

var (
	_        repository.ActivityRepository = (*fakeActivityRepo)(nil)
	fixedNow                               = time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)
)

func newTestActivityService(
	products *fakeProductRepo,
	activities *fakeActivityRepo,
	now time.Time,
) *ActivityService {
	return NewActivityService(products, activities, func() time.Time {
		return now
	})
}

func newFakeActivityRepo() *fakeActivityRepo {
	return &fakeActivityRepo{byID: map[uint64]*model.Activity{}}
}

func (f *fakeActivityRepo) Create(_ context.Context, a *model.Activity) (uint64, error) {
	f.createCalls++
	if f.createErr != nil {
		return 0, f.createErr
	}
	f.nextID++
	stored := *a
	stored.ID = f.nextID
	f.byID[stored.ID] = &stored
	return stored.ID, nil
}

func (f *fakeActivityRepo) GetByID(_ context.Context, id uint64) (*model.Activity, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.byID[id], nil
}

func seedProductWithPrice(repo *fakeProductRepo, id uint64, price uint32) {
	repo.byID[id] = &model.Product{
		ID:     id,
		Name:   "demo",
		Price:  price,
		Status: model.StatusOnSale,
	}
}

func validActivityInput() CreateActivityInput {
	start := fixedNow.Add(time.Hour)
	return CreateActivityInput{
		ProductID:    1,
		SeckillStock: 10,
		SeckillPrice: 500,
		StartTime:    start,
		EndTime:      start.Add(2 * time.Hour),
		Status:       model.ActivityStatusPending,
	}
}

func TestActivityCreateSuccess(t *testing.T) {
	products := newFakeProductRepo()
	seedProductWithPrice(products, 1, 1000)
	activities := newFakeActivityRepo()
	svc := newTestActivityService(products, activities, fixedNow)

	id, err := svc.Create(context.Background(), validActivityInput())
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if id == 0 {
		t.Fatalf("Create() id = 0, want non-zero")
	}
	if activities.createCalls != 1 {
		t.Errorf("repo.Create calls = %d, want 1", activities.createCalls)
	}
}

func TestActivityCreateInvalidParam(t *testing.T) {
	mutate := func(f func(in *CreateActivityInput)) CreateActivityInput {
		in := validActivityInput()
		f(&in)
		return in
	}

	cases := []struct {
		name string
		in   CreateActivityInput
	}{
		{"zero stock", mutate(func(in *CreateActivityInput) { in.SeckillStock = 0 })},
		{"zero price", mutate(func(in *CreateActivityInput) { in.SeckillPrice = 0 })},
		{"price not lower than product", mutate(func(in *CreateActivityInput) { in.SeckillPrice = 1000 })},
		{"zero start time", mutate(func(in *CreateActivityInput) { in.StartTime = time.Time{} })},
		{"start equals end", mutate(func(in *CreateActivityInput) { in.EndTime = in.StartTime })},
		{"start after end", mutate(func(in *CreateActivityInput) { in.EndTime = in.StartTime.Add(-time.Minute) })},
		{"bad status", mutate(func(in *CreateActivityInput) { in.Status = model.ActivityStatus(9) })},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			products := newFakeProductRepo()
			seedProductWithPrice(products, 1, 1000)
			activities := newFakeActivityRepo()
			svc := newTestActivityService(products, activities, fixedNow)

			_, err := svc.Create(context.Background(), tc.in)
			if got := appErrCode(t, err); got != errs.CodeActivityInvalid {
				t.Errorf("Create() code = %d, want %d", got, errs.CodeActivityInvalid)
			}
			if activities.createCalls != 0 {
				t.Errorf("repo.createCalls = %d, want 0", activities.createCalls)
			}
		})
	}
}

func TestActivityCreateStatusTime(t *testing.T) {
	start := fixedNow.Add(time.Hour)
	end := start.Add(2 * time.Hour)
	base := validActivityInput()

	cases := []struct {
		name    string
		now     time.Time
		status  model.ActivityStatus
		wantErr bool
	}{
		{"pending future", fixedNow, model.ActivityStatusPending, false},
		{"pending at start", start, model.ActivityStatusPending, true},
		{"pending after start", start.Add(time.Minute), model.ActivityStatusPending, true},
		{"active at start", start, model.ActivityStatusActive, false},
		{"active inside", start.Add(time.Hour), model.ActivityStatusActive, false},
		{"active before start", fixedNow, model.ActivityStatusActive, true},
		{"active at end", end, model.ActivityStatusActive, true},
		{"completed at end", end, model.ActivityStatusCompleted, false},
		{"completed after end", end.Add(time.Minute), model.ActivityStatusCompleted, false},
		{"completed before end", end.Add(-time.Minute), model.ActivityStatusCompleted, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			products := newFakeProductRepo()
			seedProductWithPrice(products, 1, 1000)
			activities := newFakeActivityRepo()
			svc := newTestActivityService(products, activities, tc.now)

			in := base
			in.Status = tc.status
			_, err := svc.Create(context.Background(), in)

			if tc.wantErr {
				if got := appErrCode(t, err); got != errs.CodeActivityInvalid {
					t.Errorf("Create() code = %d, want %d", got, errs.CodeActivityInvalid)
				}
				if activities.createCalls != 0 {
					t.Errorf("repo.createCalls = %d, want 0", activities.createCalls)
				}
				return
			}

			if err != nil {
				t.Fatalf("Create() error = %v, want nil", err)
			}
			if activities.createCalls != 1 {
				t.Errorf("repo.createCalls = %d, want 1", activities.createCalls)
			}
		})
	}
}

func TestActivityCreateProductOffShelf(t *testing.T) {
	products := newFakeProductRepo()
	seedProductWithPrice(products, 1, 1000)
	products.byID[1].Status = model.StatusOffShelf
	activities := newFakeActivityRepo()
	svc := newTestActivityService(products, activities, fixedNow)

	_, err := svc.Create(context.Background(), validActivityInput())
	if got := appErrCode(t, err); got != errs.CodeActivityInvalid {
		t.Errorf("Create() code = %d, want %d", got, errs.CodeActivityInvalid)
	}
	if activities.createCalls != 0 {
		t.Errorf("repo.createCalls = %d, want 0", activities.createCalls)
	}
}

func TestActivityCreateProductNotFound(t *testing.T) {
	products := newFakeProductRepo()
	activities := newFakeActivityRepo()
	svc := newTestActivityService(products, activities, fixedNow)

	in := validActivityInput()
	in.ProductID = 999

	_, err := svc.Create(context.Background(), in)
	if got := appErrCode(t, err); got != errs.CodeProductNotFound {
		t.Errorf("Create() code = %d, want %d", got, errs.CodeProductNotFound)
	}
	if activities.createCalls != 0 {
		t.Errorf("repo.createCalls = %d, want 0", activities.createCalls)
	}
}

func TestActivityCreateRepoError(t *testing.T) {
	products := newFakeProductRepo()
	seedProductWithPrice(products, 1, 1000)
	repoErr := errors.New("insert failed")
	activities := newFakeActivityRepo()
	activities.createErr = repoErr
	svc := newTestActivityService(products, activities, fixedNow)

	_, err := svc.Create(context.Background(), validActivityInput())
	if !errors.Is(err, repoErr) {
		t.Fatalf("Create() error = %v, want %v", err, repoErr)
	}
}

func TestActivityGetByIDNotFound(t *testing.T) {
	svc := newTestActivityService(newFakeProductRepo(), newFakeActivityRepo(), fixedNow)

	_, err := svc.GetByID(context.Background(), 999)
	if got := appErrCode(t, err); got != errs.CodeActivityNotFound {
		t.Errorf("GetByID() code = %d, want %d", got, errs.CodeActivityNotFound)
	}
}

func TestActivityGetByIDSuccess(t *testing.T) {
	activities := newFakeActivityRepo()
	activities.byID[1] = &model.Activity{ID: 1, ProductID: 1}
	svc := newTestActivityService(newFakeProductRepo(), activities, fixedNow)

	a, err := svc.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByID() error = %v, want nil", err)
	}
	if a == nil || a.ID != 1 {
		t.Fatalf("GetByID() = %+v, want id 1", a)
	}
}
