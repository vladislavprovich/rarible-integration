package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	logger2 "github.com/vladislavprovich/rarible-integration/pkg/logger"

	"github.com/unrolled/render"

	"github.com/vladislavprovich/rarible-integration/internal/handler"
	"github.com/vladislavprovich/rarible-integration/internal/service"
	"github.com/vladislavprovich/rarible-integration/pkg/client/rarible"
)

func main() {
	ctx := context.Background()
	cfg := initConfig(ctx)
	logger, err := logger2.New(ctx, cfg.Logger)
	if err != nil {
		log.Fatal(err)
	}

	baseClient := initBasicClient(ctx, logger.Logger, cfg)
	srv := initService(ctx, logger.Logger, baseClient)

	rend := render.New()
	serviceHandler := initServiceHandler(ctx, srv, logger.Logger, cfg, rend)
	router := handler.NewRouter(serviceHandler, logger.Logger, &cfg.Server)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router,
		ReadHeaderTimeout: cfg.Server.Timeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}

	go func() {
		logger.InfoContext(ctx, "Server start. Listening on port", slog.Any("port", cfg.Server.Port))
		if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("could not listen on port %s: %s", cfg.Server.Port, err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second)
	defer shutdownCancel()

	if err = httpServer.Shutdown(shutdownCtx); err != nil {
		logger.InfoContext(ctx, "Server shutdown error", slog.Any("error", err))
	}

	logger.InfoContext(ctx, "Server gracefully shutdown")
}

func initConfig(ctx context.Context) *Config {
	cfg, err := LoadConfig(ctx)
	if err != nil {
		log.Fatalf("config load error %s", err)
	}

	return cfg
}

func initBasicClient(ctx context.Context, logger *slog.Logger, cfg *Config) *rarible.BasicClient {
	logger.InfoContext(ctx, "initializing basic client")
	httpClient := &http.Client{
		Timeout: cfg.Server.HTTPClientTimeout,
	}

	baseClient := rarible.NewBasicClient(httpClient, &cfg.Client, logger)

	return baseClient
}

func initService(ctx context.Context, logger *slog.Logger, basicClient *rarible.BasicClient) *service.Service {
	logger.InfoContext(ctx, "initializing service")
	srv := service.NewRaribleService(ctx, logger, basicClient)

	return srv
}

func initServiceHandler(
	ctx context.Context,
	srv *service.Service,
	logger *slog.Logger,
	cfg *Config,
	render *render.Render,
) *handler.ServiceHandler {
	logger.InfoContext(ctx, "initializing service handler")
	serviceHandler := handler.NewServiceHandler(*srv, logger, &cfg.Server, render)

	return serviceHandler
}
