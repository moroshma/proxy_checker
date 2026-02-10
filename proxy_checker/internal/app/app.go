package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/moroshma/proxy_checker/proxy_checker/internal/config"
	"github.com/moroshma/proxy_checker/proxy_checker/internal/delivery"
	"github.com/moroshma/proxy_checker/proxy_checker/internal/repository/postgres"
	"github.com/moroshma/proxy_checker/proxy_checker/internal/service"
)

func Run(ctx context.Context, cfg *config.Config) error {
	conn, err := pgxpool.New(ctx, fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Database.User, cfg.Database.Pass, cfg.Database.Host, cfg.Database.Port, cfg.Database.DatabaseName))
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer conn.Close()

	done := make(chan error, 1)
	defer close(done)

	err = tryPingPostgres(conn)
	if err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	err = mustMigrate(cfg, conn)
	if err != nil {
		return err
	}

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	registerApi(cfg, conn, router)

	srv := initHttpServer(cfg, router)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			done <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("Shutting down the server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error(fmt.Sprintf("HTTP server shutdown error: %v", err))
		}

	case err := <-done:
		if err != nil {
			slog.Error(fmt.Sprintf("Error in runtime: %v", err))
		}
	}

	return nil
}

func initHttpServer(cfg *config.Config, r *gin.Engine) *http.Server {
	srv := &http.Server{
		Addr:         cfg.HTTP.Host + ":" + cfg.HTTP.Port,
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		Handler:      r,
	}
	return srv
}

func registerApi(cfg *config.Config, conn *pgxpool.Pool, r *gin.Engine) {
	proxyRepository := postgres.NewProxyRepository(conn)
	proxyService := service.NewResumeService(proxyRepository)
	discountHandler := delivery.NewProxyHandler(proxyService)
	delivery.RegisterServiceRoutes(r, discountHandler)

	cronChecker := service.NewCroneChecker(proxyRepository, cfg.Proxy.Timeout)
	go cronChecker.Run()
}

// mustMigrate - функция миграции базы данных
func mustMigrate(cfg *config.Config, conn *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(conn)
	driver, err := pgx.WithInstance(db, &pgx.Config{})
	defer func(driver database.Driver) {
		err := driver.Close()
		if err != nil {
			slog.Error(fmt.Sprintf("unable to close driver: %v", err))
		}
	}(driver)

	if err != nil {
		return fmt.Errorf("unable to create database driver: %w", err)
	}

	databaseInstance, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		cfg.Database.Host, driver)
	if err != nil {
		return fmt.Errorf("unable to create migration instance: %w", err)
	}

	if err := databaseInstance.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("unable to migrate database: %w", err)
	}

	version, dirty, err := databaseInstance.Version()
	if err != nil {
		return fmt.Errorf("unable to get version: %w", err)
	}
	slog.Info(fmt.Sprintf("Version: %v, Dirty: %v\n", version, dirty))
	return nil
}

// tryPingPostgres  функция пинга базы данных
func tryPingPostgres(conn *pgxpool.Pool) error {
	var err error
	for i := 0; i < 5; i++ {
		err = conn.Ping(context.Background())
		if err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("unable to ping postgres")
}
