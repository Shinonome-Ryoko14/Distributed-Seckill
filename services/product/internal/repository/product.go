// Package repository 提供 product 领域的数据访问实现。
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
)

// ProductRepository 定义商品数据访问接口，便于 service 层依赖替换与单测。
type ProductRepository interface {
	Create(ctx context.Context, p *model.Product) (uint64, error)
	GetByID(ctx context.Context, id uint64) (*model.Product, error)
	List(ctx context.Context, limit, offset int) ([]*model.Product, error)
	Count(ctx context.Context) (int64, error)
}

// MySQLProductRepository 是基于 database/sql 的 ProductRepository 实现。
type MySQLProductRepository struct {
	db *sql.DB
}

// NewMySQLProductRepository 构造 MySQLProductRepository。
func NewMySQLProductRepository(db *sql.DB) *MySQLProductRepository {
	return &MySQLProductRepository{db: db}
}

// Create 插入新商品并返回自增 ID。
func (r *MySQLProductRepository) Create(ctx context.Context, p *model.Product) (uint64, error) {
	res, err := r.db.ExecContext(ctx, "INSERT INTO products(name, description, price, status) VALUES(?,?,?,?)",
		p.Name, p.Description, p.Price, p.Status)
	if err != nil {
		return 0, fmt.Errorf("create product: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("create product id: %w", err)
	}
	return uint64(id), nil
}

// GetByID 按主键查询商品；未找到时返回 (nil, nil)。
func (r *MySQLProductRepository) GetByID(ctx context.Context, id uint64) (*model.Product, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, name, description, price, status, created_at, updated_at FROM products WHERE id = ?",
		id)

	var p model.Product
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get product by id: %w", err)
	}
	return &p, nil
}

// List 按 id 升序分页返回商品。
func (r *MySQLProductRepository) List(ctx context.Context, limit, offset int) ([]*model.Product, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, description, price, status, created_at, updated_at FROM products ORDER BY id LIMIT ? OFFSET ?",
		limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer func() { _ = rows.Close() }() // 迭代错误由上方 rows.Err() 兜底，此处安全忽略 Close 错误

	items := make([]*model.Product, 0)
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		items = append(items, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate products: %w", err)
	}
	return items, nil
}

// Count 返回商品总数，用于分页计算。
func (r *MySQLProductRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products").
		Scan(&total); err != nil {
		return 0, fmt.Errorf("count products: %w", err)
	}
	return total, nil
}
