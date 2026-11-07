package rediscli

import (
	"testing"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/config"
)

func TestNewDisablesRetriesAndSetsIOTimeout(t *testing.T) {
	client := New(config.RedisConfig{Addr: "127.0.0.1:6379"})
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close client: %v", err)
		}
	})

	opt := client.Options()

	// NewClient 会把 -1 归一化成 0；这里的 0 表示“只尝试一次”，
	// 不是 Options 默认值 3 次重试。
	if opt.MaxRetries != 0 {
		t.Errorf("MaxRetries = %d, want 0 (retries disabled)", opt.MaxRetries)
	}
	if opt.DialTimeout != ioTimeout {
		t.Errorf("DialTimeout = %s, want %s", opt.DialTimeout, ioTimeout)
	}
	if opt.ReadTimeout != ioTimeout {
		t.Errorf("ReadTimeout = %s, want %s", opt.ReadTimeout, ioTimeout)
	}
	if opt.WriteTimeout != ioTimeout {
		t.Errorf("WriteTimeout = %s, want %s", opt.WriteTimeout, ioTimeout)
	}
}
