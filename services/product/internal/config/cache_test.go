package config

import (
	"testing"
	"time"
)

func TestLoadCacheDefaults(t *testing.T) {
	t.Setenv("PRODUCT_CACHE_TTL", "")
	t.Setenv("PRODUCT_CACHE_NULL_TTL", "")

	cfg, err := LoadCache()
	if err != nil {
		t.Fatalf("LoadCache() error = %v", err)
	}
	if cfg.TTL != DefaultCacheTTL || cfg.NullTTL != DefaultCacheNullTTL {
		t.Fatalf("LoadCache() = %+v, want ttl=%s null=%s", cfg, DefaultCacheTTL, DefaultCacheNullTTL)
	}
}

func TestLoadCacheEnvOverride(t *testing.T) {
	t.Setenv("PRODUCT_CACHE_TTL", "30s")
	t.Setenv("PRODUCT_CACHE_NULL_TTL", "10s")

	cfg, err := LoadCache()
	if err != nil {
		t.Fatalf("LoadCache() err = %v", err)
	}
	if cfg.TTL != 30*time.Second || cfg.NullTTL != 10*time.Second {
		t.Errorf("LoadCache() = %+v, want ttl=30s null=10s", cfg)
	}
}

func TestLoadCacheRejectsInvalid(t *testing.T) {
	cases := map[string]struct {
		ttl     string
		nullTTL string
	}{
		"bad ttl":       {"forever", "1m"},
		"zero ttl":      {"0s", "1m"},
		"zero null ttl": {"5m", "0s"},
		"null over ttl": {"5m", "10m"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Setenv("PRODUCT_CACHE_TTL", tc.ttl)
			t.Setenv("PRODUCT_CACHE_NULL_TTL", tc.nullTTL)
			if _, err := LoadCache(); err == nil {
				t.Fatalf("LoadCache() should fail for %s", name)
			}
		})
	}
}
