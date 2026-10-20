// Package rediscli 创建 Redis 客户端并提供连通性探测。
package rediscli

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/config"
)

// ioTimeout 限制单次 Redis IO，避免缓存故障时请求被长时间拖住。
const ioTimeout = 300 * time.Millisecond

// New 构造 Redis 客户端；连接在首次命令时惰性建立，调用方负责 Close。
func New(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  ioTimeout,
		WriteTimeout: ioTimeout,
	})
}

// Ping 做一次连通性探测；失败只用于启动告警，不阻断服务启动。
func Ping(ctx context.Context, client *redis.Client) error {
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	return nil
}
