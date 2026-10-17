package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
)

// ActivityRepository 定义秒杀活动数据访问接口，便于 service 层依赖替换与单测。
type ActivityRepository interface {
	Create(ctx context.Context, a *model.Activity) (uint64, error)
	GetByID(ctx context.Context, id uint64) (*model.Activity, error)
}

// MySQLActivityRepository 是基于 database/sql 的 ActivityRepository 实现。
type MySQLActivityRepository struct {
	db *sql.DB
}

// NewMySQLActivityRepository 构造 MySQLActivityRepository。
func NewMySQLActivityRepository(db *sql.DB) *MySQLActivityRepository {
	return &MySQLActivityRepository{db: db}
}

// Create 插入新活动并返回自增 ID。
func (r *MySQLActivityRepository) Create(ctx context.Context, a *model.Activity) (uint64, error) {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO seckill_activities(product_id, seckill_stock, seckill_price, start_time, end_time, status) VALUES(?, ?, ?, ?, ?, ?)",
		a.ProductID, a.SeckillStock, a.SeckillPrice, a.StartTime, a.EndTime, a.Status)
	if err != nil {
		return 0, fmt.Errorf("create activity: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("create activity: %w", err)
	}
	return uint64(id), nil
}

// GetByID 按主键查询活动；未找到时返回 (nil, nil)。
func (r *MySQLActivityRepository) GetByID(ctx context.Context, id uint64) (*model.Activity, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, product_id, seckill_stock, seckill_price, start_time, end_time, status, created_at, updated_at FROM seckill_activities WHERE id = ?",
		id)
	var a model.Activity
	err := row.Scan(&a.ID, &a.ProductID, &a.SeckillStock, &a.SeckillPrice, &a.StartTime, &a.EndTime, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get activity by id: %w", err)
	}
	return &a, nil
}
