package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/moroshma/proxy_checker/proxy_checker/internal/app"
	"github.com/moroshma/proxy_checker/proxy_checker/internal/config"
	"github.com/moroshma/proxy_checker/proxy_checker/internal/logger"
)

func main() {
	cfg := config.MustLoad()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.New(cfg.Logger.Level)
	err := app.Run(ctx, cfg)
	if err != nil {
		slog.Error(fmt.Sprintf("error: %v", err))
	}
}
