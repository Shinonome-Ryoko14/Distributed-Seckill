// Package config 从环境变量与可选 .env 文件加载各服务共享的配置。
package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	defaultHTTPAddr          = ":8080"
	defaultLogFormat         = "json"
	defaultLogLevel          = "info"
	defaultMySQLHost         = "127.0.0.1"
	defaultMySQLPort         = "3307"
	defaultMySQLDatabase     = "seckill"
	defaultMySQLUser         = "root"
	defaultMaxOpenConns      = 20
	defaultMaxIdleConns      = 10
	defaultConnMaxLife       = time.Hour
	defaultJWTTTL            = 24 * time.Hour
	defaultMySQLDialTimeout  = 3 * time.Second
	defaultMySQLReadTimeout  = 5 * time.Second
	defaultMySQLWriteTimeout = 5 * time.Second
	defaultJWTIssuer         = "distributed-seckill"
	defaultRedisAddr         = "127.0.0.1:6379"
	defaultRedisDB           = 0
)

// Config 是全部服务的共享配置，具体服务通过环境变量覆盖默认值。
type Config struct {
	HTTP  HTTPConfig
	Log   LogConfig
	MySQL MySQLConfig
	Redis RedisConfig
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

// RedisConfig 是 Redis 客户端配置。
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
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
	dialTimeout, err := envDuration(lookup, "MYSQL_DIAL_TIMEOUT", defaultMySQLDialTimeout)
	if err != nil {
		return nil, err
	}
	readTimeout, err := envDuration(lookup, "MYSQL_READ_TIMEOUT", defaultMySQLReadTimeout)
	if err != nil {
		return nil, err
	}
	writeTimeout, err := envDuration(lookup, "MYSQL_WRITE_TIMEOUT", defaultMySQLWriteTimeout)
	if err != nil {
		return nil, err
	}

	if dialTimeout <= 0 {
		return nil, fmt.Errorf("MYSQL_DIAL_TIMEOUT must be > 0, got %s", dialTimeout)
	}
	if readTimeout <= 0 {
		return nil, fmt.Errorf("MYSQL_READ_TIMEOUT must be > 0, got %s", readTimeout)
	}
	if writeTimeout <= 0 {
		return nil, fmt.Errorf("MYSQL_WRITE_TIMEOUT must be > 0, got %s", writeTimeout)
	}

	redisDB, err := envInt(lookup, "REDIS_DB", defaultRedisDB)
	if err != nil {
		return nil, err
	}
	if redisDB < 0 {
		return nil, fmt.Errorf("REDIS_DB must be >= 0, got %d", redisDB)
	}
	jwtTTL, err := envDuration(lookup, "JWT_TTL", defaultJWTTTL)
	if err != nil {
		return nil, err
	}

	secret := lookup("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	dsn, err := mysqlDSN(lookup, dialTimeout, readTimeout, writeTimeout)
	if err != nil {
		return nil, fmt.Errorf("build mysql dsn: %w", err)
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
			DSN:             dsn,
			MaxOpenConns:    maxOpen,
			MaxIdleConns:    maxIdle,
			ConnMaxLifetime: connMaxLife,
		},
		Redis: RedisConfig{
			Addr:     env(lookup, "REDIS_ADDR", defaultRedisAddr),
			Password: lookup("REDIS_PASSWORD"),
			DB:       redisDB,
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

func mysqlDSN(lookup func(string) string, dialTimeout, readTimeout, writeTimeout time.Duration) (string, error) {
	if raw := lookup("MYSQL_DSN"); raw != "" {
		raw = strings.TrimPrefix(raw, "mysql://")

		cfg, err := mysql.ParseDSN(raw)
		if err != nil {
			return "", fmt.Errorf("MYSQL_DSN is invalid: %w", err)
		}
		if cfg.DBName == "" {
			return "", fmt.Errorf("MYSQL_DSN must include database name")
		}

		applyDefaultTimeouts(cfg, dialTimeout, readTimeout, writeTimeout)
		cfg.ParseTime = true
		return cfg.FormatDSN(), nil
	}

	password := lookup("MYSQL_PASSWORD")
	if password == "" {
		password = lookup("MYSQL_ROOT_PASSWORD")
	}

	cfg := mysql.NewConfig()
	cfg.User = env(lookup, "MYSQL_USER", defaultMySQLUser)
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(
		env(lookup, "MYSQL_HOST", defaultMySQLHost),
		env(lookup, "MYSQL_PORT", defaultMySQLPort),
	)
	cfg.DBName = env(lookup, "MYSQL_DATABASE", defaultMySQLDatabase)
	cfg.ParseTime = true
	cfg.Loc = time.Local
	cfg.Timeout = dialTimeout
	cfg.ReadTimeout = readTimeout
	cfg.WriteTimeout = writeTimeout
	cfg.Params = map[string]string{"charset": "utf8mb4"}

	return cfg.FormatDSN(), nil
}

func applyDefaultTimeouts(
	cfg *mysql.Config,
	dialTimeout, readTimeout, writeTimeout time.Duration,
) {
	if cfg.Timeout == 0 {
		cfg.Timeout = dialTimeout
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = readTimeout
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = writeTimeout
	}
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
