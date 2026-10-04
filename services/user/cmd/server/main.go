// Command server 是 user 服务的 HTTP 入口。
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/config"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/db"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/jwt"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/log"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/user/internal/handler"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/user/internal/repository"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/user/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("user service exited", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	slog.SetDefault(log.New(cfg.Log))

	conn, err := db.Open(context.Background(), cfg.MySQL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = conn.Close() }()

	tokens := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.TTL,
		cfg.JWT.Issuer)
	repo := repository.NewMySQLUserRepository(conn)
	svc := service.NewUserService(repo, tokens)
	router := handler.NewRouter(handler.NewUserHandler(svc))

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err,
			http.ErrServerClosed) {
			errCh <- err
		}
	}()
	slog.Info("user service listening", "addr", cfg.HTTP.Addr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	case sig := <-stop:
		slog.Info("shutting down", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(),
		10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
