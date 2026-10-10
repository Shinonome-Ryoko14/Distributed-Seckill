// Package model 定义 product 领域的实体与数据结构。
package model

import "time"

// Status 是商品状态码，用 int8 紧凑存储并预留扩展值。
type Status int8

// 商品状态取值，与 products.status（TINYINT）对应。
const (
	StatusOnSale   Status = 1
	StatusOffShelf Status = 2
)

// Product 是商品领域实体；Price 以分存储避免浮点误差，状态受 Status 枚举约束。
type Product struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       uint32    `json:"price"` // 单位：分
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
