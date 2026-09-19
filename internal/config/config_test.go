package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func fakeLookup(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := load(fakeLookup(map[string]string{"JWT_SECRET": "s3cret"}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.HTTP.Addr != defaultHTTPAddr {
		t.Errorf("HTTP.Addr = %q, want %q", cfg.HTTP.Addr, defaultHTTPAddr)
	}
	if cfg.Log.Format != defaultLogFormat || cfg.Log.Level != defaultLogLevel {
		t.Errorf("Log = %+v, want format=%q level=%q", cfg.Log, defaultLogFormat, defaultLogLevel)
	}
	wantDSN := "root:@tcp(127.0.0.1:3307)/seckill?charset=utf8mb4&parseTime=true&loc=Local"
	if cfg.MySQL.DSN != wantDSN {
		t.Errorf("MySQL.DSN = %q, want %q", cfg.MySQL.DSN, wantDSN)
	}
	if cfg.MySQL.MaxOpenConns != defaultMaxOpenConns || cfg.MySQL.MaxIdleConns != defaultMaxIdleConns {
		t.Errorf("MySQL pool = (%d,%d), want (%d,%d)",
			cfg.MySQL.MaxOpenConns, cfg.MySQL.MaxIdleConns, defaultMaxOpenConns, defaultMaxIdleConns)
	}
	if cfg.JWT.Secret != "s3cret" || cfg.JWT.TTL != defaultJWTTTL || cfg.JWT.Issuer != defaultJWTIssuer {
		t.Errorf("JWT = %+v, want secret=s3cret ttl=%s issuer=%q", cfg.JWT, defaultJWTTTL, defaultJWTIssuer)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	cfg, err := load(fakeLookup(map[string]string{
		"JWT_SECRET":              "s",
		"HTTP_ADDR":               ":9000",
		"MYSQL_DSN":               "custom-dsn",
		"JWT_TTL":                 "30m",
		"MYSQL_MAX_OPEN_CONNS":    "5",
		"MYSQL_MAX_IDLE_CONNS":    "3",
		"MYSQL_CONN_MAX_LIFETIME": "5m",
		"LOG_FORMAT":              "TEXT",
		"LOG_LEVEL":               "DEBUG",
	}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.HTTP.Addr != ":9000" {
		t.Errorf("HTTP.Addr = %q, want :9000", cfg.HTTP.Addr)
	}
	if cfg.MySQL.DSN != "custom-dsn" {
		t.Errorf("MySQL.DSN = %q, want custom-dsn", cfg.MySQL.DSN)
	}
	if cfg.JWT.TTL != 30*time.Minute {
		t.Errorf("JWT.TTL = %s, want 30m", cfg.JWT.TTL)
	}
	if cfg.MySQL.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("MySQL.ConnMaxLifetime = %s, want 5m", cfg.MySQL.ConnMaxLifetime)
	}
	if cfg.Log.Format != "text" || cfg.Log.Level != "debug" {
		t.Errorf("Log = %+v, want format=text level=debug", cfg.Log)
	}
}

func TestLoadPasswordFallback(t *testing.T) {
	cfg, err := load(fakeLookup(map[string]string{
		"JWT_SECRET":          "s",
		"MYSQL_ROOT_PASSWORD": "rootpw",
	}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	want := "root:rootpw@tcp(127.0.0.1:3307)/seckill?charset=utf8mb4&parseTime=true&loc=Local"
	if cfg.MySQL.DSN != want {
		t.Errorf("MySQL.DSN = %q, want %q", cfg.MySQL.DSN, want)
	}
}

func TestLoadRequiresJWTSecret(t *testing.T) {
	if _, err := load(fakeLookup(nil)); err == nil {
		t.Fatal("load should fail when JWT_SECRET is empty")
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	cases := map[string]map[string]string{
		"bad int":      {"JWT_SECRET": "s", "MYSQL_MAX_OPEN_CONNS": "abc"},
		"bad duration": {"JWT_SECRET": "s", "JWT_TTL": "forever"},
		"idle > open":  {"JWT_SECRET": "s", "MYSQL_MAX_OPEN_CONNS": "2", "MYSQL_MAX_IDLE_CONNS": "3"},
	}
	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := load(fakeLookup(env)); err == nil {
				t.Fatalf("load should fail for %s", name)
			}
		})
	}
}

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "# comment\n\nDS_TEST_PLAIN=bar\nexport DS_TEST_QUOTED=\"baz qux\"\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	const plain, quoted = "DS_TEST_PLAIN", "DS_TEST_QUOTED"
	t.Cleanup(func() {
		_ = os.Unsetenv(plain)
		_ = os.Unsetenv(quoted)
	})

	if err := LoadDotEnv(path); err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if got := os.Getenv(plain); got != "bar" {
		t.Errorf("%s = %q, want bar", plain, got)
	}
	if got := os.Getenv(quoted); got != "baz qux" {
		t.Errorf("%s = %q, want %q", quoted, got, "baz qux")
	}
}

func TestLoadDotEnvDoesNotOverride(t *testing.T) {
	const key = "DS_TEST_EXISTING"
	t.Setenv(key, "original")

	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(key+"=fromfile\n"), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
	if err := LoadDotEnv(path); err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if got := os.Getenv(key); got != "original" {
		t.Errorf("%s = %q, want original", key, got)
	}
}

func TestLoadDotEnvMissingFile(t *testing.T) {
	if err := LoadDotEnv(filepath.Join(t.TempDir(), "nope.env")); err != nil {
		t.Fatalf("missing file should be ignored, got %v", err)
	}
}
