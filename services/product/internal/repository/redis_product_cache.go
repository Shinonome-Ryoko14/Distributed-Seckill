package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/model"
)

const productCacheKeyPrefix = "product:detail:"

// redis.Client又是什么作用？

// RedisProductCache 是基于 Redis 的 ProductCache 实现。
// 负缓存以 JSON null 存储；正常值存储 model.Product 的 JSON。
type RedisProductCache struct {
	client *redis.Client
}

var _ ProductCache = (*RedisProductCache)(nil)

// NewRedisProductCache 构造 RedisProductCache。
func NewRedisProductCache(client *redis.Client) *RedisProductCache {
	return &RedisProductCache{client: client}
}

// Get 读取缓存；miss 与负缓存通过 hit 区分。
func (c *RedisProductCache) Get(ctx context.Context, id uint64) (*model.Product, bool, error) {
	key := productCacheKey(id)

	// 这里的自调用递归是什么意思？
	raw, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var p *model.Product
	// 为什么反序列化成功就相当于一次hit？什么叫脏数据影响？
	// 反序列化到 model.Product，才能把 JSON null 还原成 nil。
	if err := json.Unmarshal(raw, &p); err != nil {
		// 脏数据不应永久阻塞读取：尽力删除，按未命中回源。
		_ = c.client.Del(ctx, key).Err()
		return nil, false, nil
	}
	return p, true, nil
}

// Set 写入缓存；p == nil 时写入负缓存。
func (c *RedisProductCache) Set(ctx context.Context, id uint64, p *model.Product, ttl time.Duration) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	// 为什么这里的传参和product_cache.go中设置的参数类型不一样？(比如第三个类型是interface，但是传入[]byte)
	return c.client.Set(ctx, productCacheKey(id), raw, ttl).Err()
}

// Delete 删除缓存；key 不存在时也返回 nil。
func (c *RedisProductCache) Delete(ctx context.Context, id uint64) error {
	return c.client.Del(ctx, productCacheKey(id)).Err()
}

func productCacheKey(id uint64) string {
	return productCacheKeyPrefix + strconv.FormatUint(id, 10)
}
