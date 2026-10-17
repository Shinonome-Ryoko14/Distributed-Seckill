// Package model 定义 product 领域的实体与数据结构。
package model

import "time"

// ActivityStatus 是秒杀活动状态码，与 seckill_activities.status（TINYINT）对应。
type ActivityStatus int8

// 活动状态取值。
const (
	ActivityStatusPending   ActivityStatus = 1
	ActivityStatusActive    ActivityStatus = 2
	ActivityStatusCompleted ActivityStatus = 3
)

// Activity 是秒杀活动领域实体。
type Activity struct {
	ID           uint64         `json:"id"`
	ProductID    uint64         `json:"product_id"`
	SeckillStock uint32         `json:"seckill_stock"`
	SeckillPrice uint32         `json:"seckill_price"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	Status       ActivityStatus `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
