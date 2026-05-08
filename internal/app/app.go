package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Danil-Ivonin/TestSubs/internal/config"
	"github.com/Danil-Ivonin/TestSubs/internal/httpserver"
	"github.com/Danil-Ivonin/TestSubs/internal/logger"
	postgresdb "github.com/Danil-Ivonin/TestSubs/internal/postgres"
	"github.com/Danil-Ivonin/TestSubs/internal/subscription"
)

func Run(ctx context.Context) error {
	cfg, err := config.Load(config.Options{})
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	log, err := logger.New(cfg.Log)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgresdb.NewPool(ctx, cfg.Postgres.DSN)
	if err != nil {
		return fmt.Errorf("connecting postgres: %w", err)
	}
	defer pool.Close()

	subscriptionRepo := subscription.NewPostgresRepository(pool)
	subscriptionService := subscription.NewService(subscriptionRepo)
	subscriptionHandler := subscription.NewHandler(subscriptionService, log)

	router := httpserver.NewRouter(log, httpserver.RouterOptions{
		SubscriptionHandler: subscriptionHandler,
	})
	server := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.WithField("addr", server.Addr).Info("starting http server")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Info("shutting down http server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutting down server: %w", err)
		}
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("running server: %w", err)
		}
		return nil
	}
}
