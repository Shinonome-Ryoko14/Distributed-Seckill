package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
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
	dsn, err := mysql.ParseDSN(cfg.MySQL.DSN)
	if err != nil {
		t.Fatalf("parse MySQL.DSN: %v", err)
	}
	if dsn.User != defaultMySQLUser {
		t.Errorf("DSN user = %q, want %q", dsn.User, defaultMySQLUser)
	}
	if dsn.Passwd != "" {
		t.Errorf("DSN password = %q, want empty", dsn.Passwd)
	}
	if dsn.Addr != "127.0.0.1:3307" {
		t.Errorf("DSN addr = %q, want 127.0.0.1:3307", dsn.Addr)
	}
	if dsn.DBName != defaultMySQLDatabase {
		t.Errorf("DSN database = %q, want %q", dsn.DBName, defaultMySQLDatabase)
	}
	if !dsn.ParseTime {
		t.Errorf("DSN parseTime = false, want true")
	}
	if dsn.Loc.String() != "Local" {
		t.Errorf("DSN loc = %q, want Local", dsn.Loc.String())
	}
	if dsn.Timeout != defaultMySQLDialTimeout ||
		dsn.ReadTimeout != defaultMySQLReadTimeout ||
		dsn.WriteTimeout != defaultMySQLWriteTimeout {
		t.Errorf("DSN timeouts = (%s,%s,%s), want(%s,%s,%s)", dsn.Timeout, dsn.ReadTimeout, dsn.WriteTimeout,
			defaultMySQLDialTimeout, defaultMySQLReadTimeout, defaultMySQLWriteTimeout)
	}
	if !strings.Contains(cfg.MySQL.DSN, "charset=utf8mb4") {
		t.Errorf("DSN = %q, want charset=utf8mb4", cfg.MySQL.DSN)
	}
	if cfg.MySQL.MaxOpenConns != defaultMaxOpenConns || cfg.MySQL.MaxIdleConns != defaultMaxIdleConns {
		t.Errorf("MySQL pool = (%d,%d), want (%d,%d)",
			cfg.MySQL.MaxOpenConns, cfg.MySQL.MaxIdleConns, defaultMaxOpenConns, defaultMaxIdleConns)
	}
	if cfg.JWT.Secret != "s3cret" || cfg.JWT.TTL != defaultJWTTTL || cfg.JWT.Issuer != defaultJWTIssuer {
		t.Errorf("JWT = %+v, want secret=s3cret ttl=%s issuer=%q", cfg.JWT, defaultJWTTTL, defaultJWTIssuer)
	}
	if cfg.Redis.Addr != defaultRedisAddr || cfg.Redis.Password != "" || cfg.Redis.DB != defaultRedisDB {
		t.Errorf("Redis = %+v, want addr=%q db=%d", cfg.Redis, defaultRedisAddr, defaultRedisDB)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	cfg, err := load(fakeLookup(map[string]string{
		"JWT_SECRET":              "s",
		"HTTP_ADDR":               ":9000",
		"MYSQL_DSN":               "app:secret@tcp(db:3307)/custom?parseTime=true",
		"JWT_TTL":                 "30m",
		"MYSQL_MAX_OPEN_CONNS":    "5",
		"MYSQL_MAX_IDLE_CONNS":    "3",
		"MYSQL_CONN_MAX_LIFETIME": "5m",
		"LOG_FORMAT":              "TEXT",
		"LOG_LEVEL":               "DEBUG",
		"REDIS_ADDR":              "redis:6380",
		"REDIS_PASSWORD":          "redis-pw",
		"REDIS_DB":                "3",
	}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.HTTP.Addr != ":9000" {
		t.Errorf("HTTP.Addr = %q, want :9000", cfg.HTTP.Addr)
	}
	dsn, err := mysql.ParseDSN(cfg.MySQL.DSN)
	if err != nil {
		t.Fatalf("parse MySQL.DSN: %v", err)
	}
	if dsn.User != "app" || dsn.Passwd != "secret" ||
		dsn.Addr != "db:3307" || dsn.DBName != "custom" {
		t.Errorf("DSN = %+v, want app:secret@db:3307/custom", dsn)
	}
	if dsn.Timeout != defaultMySQLDialTimeout ||
		dsn.ReadTimeout != defaultMySQLReadTimeout ||
		dsn.WriteTimeout != defaultMySQLWriteTimeout {
		t.Errorf("DSN timeouts = (%s,%s,%s), want defaults",
			dsn.Timeout, dsn.ReadTimeout, dsn.WriteTimeout)
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
	if cfg.Redis.Addr != "redis:6380" || cfg.Redis.Password != "redis-pw" || cfg.Redis.DB != 3 {
		t.Errorf("Redis = %+v, want addr=redis:6380 password=redis-pw db=3", cfg.Redis)
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

	dsn, err := mysql.ParseDSN(cfg.MySQL.DSN)
	if err != nil {
		t.Fatalf("parse MySQL.DSN: %v", err)
	}
	if dsn.Passwd != "rootpw" {
		t.Errorf("DSN password = %q, want rootpw", dsn.Passwd)
	}
}

func TestLoadDSNOverride(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"plain", "app:secret@tcp(db:3307)/custom"},
		{"legacy scheme", "mysql://app:secret@tcp(db:3307)/custom"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := load(fakeLookup(map[string]string{
				"JWT_SECRET": "s",
				"MYSQL_DSN":  tc.raw,
			}))
			if err != nil {
				t.Fatalf("load: %v", err)
			}

			dsn, err := mysql.ParseDSN(cfg.MySQL.DSN)
			if err != nil {
				t.Fatalf("parse DSN: %v", err)
			}
			if dsn.User != "app" || dsn.Passwd != "secret" ||
				dsn.Addr != "db:3307" || dsn.DBName != "custom" {
				t.Errorf("DSN = %+v, want app:secret@db:3307/custom", dsn)
			}
			if dsn.Timeout != defaultMySQLDialTimeout ||
				dsn.ReadTimeout != defaultMySQLReadTimeout ||
				dsn.WriteTimeout != defaultMySQLWriteTimeout {
				t.Errorf("DSN timeouts not defaulted: %+v", dsn)
			}
		})
	}
}

func TestLoadDSNSpecialPassword(t *testing.T) {
	const password = "p@ss:word/with"

	cfg, err := load(fakeLookup(map[string]string{
		"JWT_SECRET":     "s",
		"MYSQL_PASSWORD": password,
	}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	dsn, err := mysql.ParseDSN(cfg.MySQL.DSN)
	if err != nil {
		t.Fatalf("parse DSN: %v", err)
	}
	if dsn.Passwd != password {
		t.Errorf("DSN password = %q, want %q", dsn.Passwd, password)
	}
}

func TestLoadRequiresJWTSecret(t *testing.T) {
	if _, err := load(fakeLookup(nil)); err == nil {
		t.Fatal("load should fail when JWT_SECRET is empty")
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	cases := map[string]map[string]string{
		"bad int":                {"JWT_SECRET": "s", "MYSQL_MAX_OPEN_CONNS": "abc"},
		"bad duration":           {"JWT_SECRET": "s", "JWT_TTL": "forever"},
		"idle > open":            {"JWT_SECRET": "s", "MYSQL_MAX_OPEN_CONNS": "2", "MYSQL_MAX_IDLE_CONNS": "3"},
		"bad redis db":           {"JWT_SECRET": "s", "REDIS_DB": "abc"},
		"negative redis db":      {"JWT_SECRET": "s", "REDIS_DB": "-1"},
		"bad mysql dsn":          {"JWT_SECRET": "s", "MYSQL_DSN": "not-a-dsn"},
		"zero dial timeout":      {"JWT_SECRET": "s", "MYSQL_DIAL_TIMEOUT": "0s"},
		"negative read timeout":  {"JWT_SECRET": "s", "MYSQL_READ_TIMEOUT": "-1s"},
		"negative write timeout": {"JWT_SECRET": "s", "MYSQL_WRITE_TIMEOUT": "-1s"},
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
