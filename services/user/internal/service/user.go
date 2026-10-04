// Package service 实现 user 领域的注册与登录业务逻辑。
package service

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/jwt"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/user/internal/model"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/user/internal/repository"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,32}$`)

func validateCredentials(username, password string) error {
	if !usernamePattern.MatchString(username) {
		return errs.New(errs.CodeInvalidParam, "invalid username",
			http.StatusBadRequest)
	}
	if len(password) < 8 || len(password) > 72 {
		return errs.New(errs.CodeInvalidParam, "invalid password",
			http.StatusBadRequest)
	}
	return nil
}

// UserService 提供用户注册与登录业务。
type UserService struct {
	repo   repository.UserRepository
	tokens *jwt.Manager
}

// NewUserService 构造 UserService，注入数据访问与 JWT 签发依赖。
func NewUserService(repo repository.UserRepository, tokens *jwt.Manager) *UserService {
	return &UserService{repo: repo, tokens: tokens}
}

// Register 校验入参、bcrypt 哈希后创建用户，返回新用户 ID。
func (s *UserService) Register(ctx context.Context, username, password string) (uint64, error) {
	err := validateCredentials(username, password)
	if err != nil {
		return 0, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password),
		bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}

	return s.repo.Create(ctx, &model.User{Username: username, PasswordHash: string(hash)})
}

// Login 校验账号密码并签发 JWT，返回 token。
func (s *UserService) Login(ctx context.Context, username, password string) (string, error) {
	u, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", errs.New(errs.CodeUserNotFound, "user not found",
			http.StatusUnauthorized)
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return "", errs.New(errs.CodePasswordWrong, "incorrect password",
			http.StatusUnauthorized)
	}

	token, err := s.tokens.Generate(u.ID)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}
