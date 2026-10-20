// Package config 加载 product 服务特有的配置。
package config

import (
	"fmt"
	"os"
	"time"
)

// 商品详情缓存默认 TTL。
const (
	DefaultCacheTTL     = 5 * time.Minute
	DefaultCacheNullTTL = time.Minute
)

// CacheConfig 是商品详情缓存配置。
type CacheConfig struct {
	TTL     time.Duration
	NullTTL time.Duration
}

// LoadCache 从环境变量加载缓存 TTL，并校验取值关系。
func LoadCache() (CacheConfig, error) {
	ttl, err := durationEnv("PRODUCT_CACHE_TTL", DefaultCacheTTL)
	if err != nil {
		return CacheConfig{}, err
	}

	nullTTL, err := durationEnv("PRODUCT_CACHE_NULL_TTL", DefaultCacheNullTTL)
	if err != nil {
		return CacheConfig{}, err
	}

	if ttl <= 0 {
		return CacheConfig{}, fmt.Errorf("PRODUCT_CACHE_TTL must be positive, got %s", ttl)
	}
	if nullTTL <= 0 {
		return CacheConfig{}, fmt.Errorf("PRODUCT_CACHE_NULL_TTL must be positive, got %s", nullTTL)
	}
	if nullTTL > ttl {
		return CacheConfig{}, fmt.Errorf("PRODUCT_CACHE_NULL_TTL (%s) must not exceed PRODUCT_CACHE_TTL (%s) ",
			nullTTL, ttl)
	}
	return CacheConfig{TTL: ttl, NullTTL: nullTTL}, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	return v, nil
}
