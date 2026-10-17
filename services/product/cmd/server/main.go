// Command server 是 product 服务的 HTTP 入口。
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
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/log"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/handler"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/repository"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/services/product/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("product service exited", "error", err)
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

	productRepo := repository.NewMySQLProductRepository(conn)
	activityRepo := repository.NewMySQLActivityRepository(conn)

	productSvc := service.NewProductService(productRepo)
	activitySvc := service.NewActivityService(productRepo, activityRepo)

	router := handler.NewRouter(
		handler.NewProductHandler(productSvc),
		handler.NewActivityHandler(activitySvc),
	)

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	slog.Info("product service listening", "addr", cfg.HTTP.Addr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	case sig := <-stop:
		slog.Info("shutting down", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
