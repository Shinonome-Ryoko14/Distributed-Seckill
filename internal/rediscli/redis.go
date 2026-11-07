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
		DialTimeout:  ioTimeout,
		ReadTimeout:  ioTimeout,
		WriteTimeout: ioTimeout,
		// -1 才是禁用重试；0 会回落到 go-redis 默认的 3 次。
		MaxRetries: -1,
	})
}

// Ping 做一次连通性探测；失败只用于启动告警，不阻断服务启动。
func Ping(ctx context.Context, client *redis.Client) error {
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	return nil
}
