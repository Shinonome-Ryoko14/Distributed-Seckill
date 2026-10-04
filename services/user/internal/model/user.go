// Package model 定义 user 领域的实体与数据结构。
package model

import "time"

// User 是用户领域实体，含密码哈希，仅供服务内部流转，不直接对外暴露。
type User struct {
	ID           uint64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
