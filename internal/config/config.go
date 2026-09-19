// Package config 从环境变量与可选 .env 文件加载各服务共享的配置。
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddr      = ":8080"
	defaultLogFormat     = "json"
	defaultLogLevel      = "info"
	defaultMySQLHost     = "127.0.0.1"
	defaultMySQLPort     = "3307"
	defaultMySQLDatabase = "seckill"
	defaultMySQLUser     = "root"
	defaultMaxOpenConns  = 20
	defaultMaxIdleConns  = 10
	defaultConnMaxLife   = time.Hour
	defaultJWTTTL        = 24 * time.Hour
	defaultJWTIssuer     = "distributed-seckill"
)

// Config 是全部服务的共享配置，具体服务通过环境变量覆盖默认值。
type Config struct {
	HTTP  HTTPConfig
	Log   LogConfig
	MySQL MySQLConfig
	JWT   JWTConfig
}

// HTTPConfig 是 HTTP 服务监听配置。
type HTTPConfig struct {
	Addr string
}

// LogConfig 是结构化日志配置。
type LogConfig struct {
	Format string
	Level  string
}

// MySQLConfig 是 MySQL 连接池配置。
type MySQLConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig 是 JWT 签发与校验配置。
type JWTConfig struct {
	Secret string
	TTL    time.Duration
	Issuer string
}

// Load 先以 best-effort 方式加载 ENV_FILE（默认 deploy/.env），再读取进程环境变量。
// 已存在的进程环境变量优先，不会被 .env 覆盖。
func Load() (*Config, error) {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = "deploy/.env"
	}
	if err := LoadDotEnv(envFile); err != nil {
		return nil, fmt.Errorf("load env file %s: %w", envFile, err)
	}
	return load(os.Getenv)
}

func load(lookup func(string) string) (*Config, error) {
	maxOpen, err := envInt(lookup, "MYSQL_MAX_OPEN_CONNS", defaultMaxOpenConns)
	if err != nil {
		return nil, err
	}
	maxIdle, err := envInt(lookup, "MYSQL_MAX_IDLE_CONNS", defaultMaxIdleConns)
	if err != nil {
		return nil, err
	}
	connMaxLife, err := envDuration(lookup, "MYSQL_CONN_MAX_LIFETIME", defaultConnMaxLife)
	if err != nil {
		return nil, err
	}
	jwtTTL, err := envDuration(lookup, "JWT_TTL", defaultJWTTTL)
	if err != nil {
		return nil, err
	}

	secret := lookup("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	cfg := &Config{
		HTTP: HTTPConfig{
			Addr: env(lookup, "HTTP_ADDR", defaultHTTPAddr),
		},
		Log: LogConfig{
			Format: strings.ToLower(env(lookup, "LOG_FORMAT", defaultLogFormat)),
			Level:  strings.ToLower(env(lookup, "LOG_LEVEL", defaultLogLevel)),
		},
		MySQL: MySQLConfig{
			DSN:             mysqlDSN(lookup),
			MaxOpenConns:    maxOpen,
			MaxIdleConns:    maxIdle,
			ConnMaxLifetime: connMaxLife,
		},
		JWT: JWTConfig{
			Secret: secret,
			TTL:    jwtTTL,
			Issuer: env(lookup, "JWT_ISSUER", defaultJWTIssuer),
		},
	}

	if cfg.MySQL.MaxIdleConns > cfg.MySQL.MaxOpenConns {
		return nil, fmt.Errorf("MYSQL_MAX_IDLE_CONNS (%d) must not exceed MYSQL_MAX_OPEN_CONNS (%d)",
			cfg.MySQL.MaxIdleConns, cfg.MySQL.MaxOpenConns)
	}
	return cfg, nil
}

func mysqlDSN(lookup func(string) string) string {
	if dsn := lookup("MYSQL_DSN"); dsn != "" {
		return dsn
	}
	password := lookup("MYSQL_PASSWORD")
	if password == "" {
		password = lookup("MYSQL_ROOT_PASSWORD")
	}
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		env(lookup, "MYSQL_USER", defaultMySQLUser),
		password,
		env(lookup, "MYSQL_HOST", defaultMySQLHost),
		env(lookup, "MYSQL_PORT", defaultMySQLPort),
		env(lookup, "MYSQL_DATABASE", defaultMySQLDatabase),
	)
}

func env(lookup func(string) string, key, fallback string) string {
	if v := lookup(key); v != "" {
		return v
	}
	return fallback
}

func envInt(lookup func(string) string, key string, fallback int) (int, error) {
	raw := lookup(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return v, nil
}

func envDuration(lookup func(string) string, key string, fallback time.Duration) (time.Duration, error) {
	raw := lookup(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	return v, nil
}
