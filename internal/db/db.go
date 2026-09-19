// Package db 用 database/sql 建立 MySQL 连接池。
package db

import (
	"context"
	"database/sql"
	"fmt"

	// 注册 database/sql 的 "mysql" 驱动。
	_ "github.com/go-sql-driver/mysql"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/config"
)

// Open 建立 MySQL 连接池并做一次连通性探测。返回的 *sql.DB 由调用方负责 Close。
func Open(ctx context.Context, cfg config.MySQLConfig) (*sql.DB, error) {
	conn, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)
	conn.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := conn.PingContext(ctx); err != nil {
		// 关闭失败无法覆盖原始错误，best-effort 释放连接池。
		_ = conn.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return conn, nil
}
