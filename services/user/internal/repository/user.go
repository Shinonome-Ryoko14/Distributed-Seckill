// Package repository 提供 user 领域的数据访问实现。
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-sql-driver/mysql"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/user/internal/model"
)

// UserRepository 定义 user 数据访问接口，便于 service 层依赖替换与单测。
type UserRepository interface {
	Create(ctx context.Context, u *model.User) (uint64, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
}

// Create 插入新用户并返回自增 ID；用户名冲突时返回 errs.CodeUserExists。
func (r *MySQLUserRepository) Create(ctx context.Context, u *model.User) (uint64, error) {
	res, err := r.db.ExecContext(ctx, "INSERT INTO users(username, password_hash) VALUES(?, ?) ",
		u.Username, u.PasswordHash)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return 0, errs.New(errs.CodeUserExists, "username already exists", http.StatusConflict)
		}
		return 0, fmt.Errorf("create user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	return uint64(id), nil
}

// GetByUsername 按用户名查询用户；未找到时返回 (nil, nil)。
func (r *MySQLUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = ?",
		username)
	var u model.User

	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return &u, nil
}

// MySQLUserRepository 是基于 database/sql 的 UserRepository 实现。
type MySQLUserRepository struct {
	db *sql.DB
}

// NewMySQLUserRepository 构造 MySQLUserRepository。
func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}
