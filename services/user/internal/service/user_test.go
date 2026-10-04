package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/jwt"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/user/internal/model"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/user/internal/repository"
)

// fakeUserRepo 是 repository.UserRepository 的内存实现，用于隔离数据库。
type fakeUserRepo struct {
	byUsername  map[string]*model.User
	nextID      uint64
	createCalls int
	createErr   error
	getErr      error
}

var _ repository.UserRepository = (*fakeUserRepo)(nil)

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byUsername: map[string]*model.User{}}
}

func (f *fakeUserRepo) Create(_ context.Context, u *model.User) (uint64, error) {
	f.createCalls++
	if f.createErr != nil {
		return 0, f.createErr
	}
	if _, ok := f.byUsername[u.Username]; ok {
		return 0, errs.New(errs.CodeUserExists, "username already exists", http.StatusConflict)
	}
	f.nextID++
	stored := *u
	stored.ID = f.nextID
	f.byUsername[stored.Username] = &stored
	return stored.ID, nil
}

func (f *fakeUserRepo) GetByUsername(_ context.Context, username string) (*model.User, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	u, ok := f.byUsername[username]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func newTestService(repo repository.UserRepository) (*UserService, *jwt.Manager) {
	manager := jwt.NewManager("test-secret", time.Hour, "test")
	return NewUserService(repo, manager), manager
}

func appErrCode(t *testing.T, err error) errs.Code {
	t.Helper()
	var appErr *errs.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %v, want *errs.AppError", err)
	}
	return appErr.Code
}

func TestRegisterSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	svc, _ := newTestService(repo)

	id, err := svc.Register(context.Background(), "alice", "secret-password")
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}
	if id == 0 {
		t.Fatal("Register() id = 0, want non-zero")
	}

	u := repo.byUsername["alice"]
	if u == nil {
		t.Fatal("registered user not found in repo")
	}
	if u.PasswordHash == "secret-password" {
		t.Error("password stored in plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("secret-password")); err != nil {
		t.Errorf("stored hash does not match password: %v", err)
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	repo := newFakeUserRepo()
	svc, _ := newTestService(repo)

	if _, err := svc.Register(context.Background(), "bob", "secret-password"); err != nil {
		t.Fatalf("first Register() error = %v, want nil", err)
	}
	_, err := svc.Register(context.Background(), "bob", "another-password")
	if got := appErrCode(t, err); got != errs.CodeUserExists {
		t.Errorf("duplicate Register() code = %d, want %d", got, errs.CodeUserExists)
	}
}

func TestRegisterInvalidParam(t *testing.T) {
	cases := []struct {
		name     string
		username string
		password string
	}{
		{"empty username", "", "secret-password"},
		{"short username", "ab", "secret-password"},
		{"illegal char", "bad name!", "secret-password"},
		{"long username", strings.Repeat("a", 33), "secret-password"},
		{"short password", "charlie", "1234567"},
		{"long password", "charlie", strings.Repeat("a", 73)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeUserRepo()
			svc, _ := newTestService(repo)

			_, err := svc.Register(context.Background(), tc.username, tc.password)
			if got := appErrCode(t, err); got != errs.CodeInvalidParam {
				t.Errorf("Register() code = %d, want %d", got, errs.CodeInvalidParam)
			}
			if repo.createCalls != 0 {
				t.Errorf("repo.Create called %d times, want 0", repo.createCalls)
			}
		})
	}
}

func TestLoginSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	svc, manager := newTestService(repo)

	id, err := svc.Register(context.Background(), "carol", "secret-password")
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	token, err := svc.Login(context.Background(), "carol", "secret-password")
	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}
	if token == "" {
		t.Fatal("Login() token is empty")
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse(token) error = %v, want nil", err)
	}
	if claims.UserID != id {
		t.Errorf("claims.UserID = %d, want %d", claims.UserID, id)
	}
}

func TestLoginUserNotFound(t *testing.T) {
	svc, _ := newTestService(newFakeUserRepo())

	_, err := svc.Login(context.Background(), "nobody", "secret-password")
	if got := appErrCode(t, err); got != errs.CodeUserNotFound {
		t.Errorf("Login() code = %d, want %d", got, errs.CodeUserNotFound)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	repo := newFakeUserRepo()
	svc, _ := newTestService(repo)

	if _, err := svc.Register(context.Background(), "dave", "secret-password"); err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	_, err := svc.Login(context.Background(), "dave", "wrong-password")
	if got := appErrCode(t, err); got != errs.CodePasswordWrong {
		t.Errorf("Login() code = %d, want %d", got, errs.CodePasswordWrong)
	}
}

func TestLoginRepoErrorPropagates(t *testing.T) {
	repo := newFakeUserRepo()
	sentinel := errors.New("db down")
	repo.getErr = sentinel
	svc, _ := newTestService(repo)

	_, err := svc.Login(context.Background(), "alice", "secret-password")
	if !errors.Is(err, sentinel) {
		t.Fatalf("Login() error = %v, want wrapping of sentinel", err)
	}
	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		t.Error("repo system error must not become a business AppError")
	}
}

func TestRegisterRepoErrorPropagates(t *testing.T) {
	repo := newFakeUserRepo()
	sentinel := errors.New("db down")
	repo.createErr = sentinel
	svc, _ := newTestService(repo)

	_, err := svc.Register(context.Background(), "erin", "secret-password")
	if !errors.Is(err, sentinel) {
		t.Fatalf("Register() error = %v, want wrapping of sentinel", err)
	}
}
