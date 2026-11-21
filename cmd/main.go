package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Rizwank123/chatApp/internal/dependency"
	"github.com/Rizwank123/chatApp/internal/http/swagger"
	"github.com/Rizwank123/chatApp/internal/pkg/config"
	"github.com/Rizwank123/chatApp/internal/service"
	"github.com/labstack/echo/v4"
)

func main() {
	cfg, err := dependency.NewConfig(getConfigOptions())
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := dependency.NewDatabaseConfig(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	handlers, err := dependency.NewChatAppApi(cfg, db)
	if err != nil {
		_ = db.Close()
		log.Fatalf("failed to initialize application dependencies: %v", err)
	}

	appCtx, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()
	hubDone := make(chan struct{})
	go func() {
		defer close(hubDone)
		handlers.Hub.Run(appCtx)
	}()

	brokerDone := make(chan struct{})
	go func() {
		defer close(brokerDone)
		if err := service.DeliverBrokerMessages(appCtx, handlers.Broker, handlers.Hub); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("message broker stopped: %v", err)
		}
	}()

	e := echo.New()
	e.HideBanner = true
	if cfg.AppEnv == "production" {
		e.HidePort = true
	}
	e.Server.ReadHeaderTimeout = 10 * time.Second
	e.Server.IdleTimeout = 120 * time.Second
	e.Server.MaxHeaderBytes = 1 << 20

	e.GET("/health/live", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/health/ready", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "database unavailable")
		}
		if err := handlers.Redis.Client().Ping(ctx).Err(); err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "cache unavailable")
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	})

	handlers.API.SetupMiddleware(e)
	swagger.SetupSwagger(cfg, e)
	handlers.API.SetupRoutes(e)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.AppPort))
	}()

	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	select {
	case <-signalCtx.Done():
		log.Print("shutdown signal received")
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server stopped: %v", err)
		}
	}

	// Stop accepting requests and cancel long-lived background connections before
	// closing the shared dependencies they use.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	cancelApp()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown: %v", err)
	}
	select {
	case <-brokerDone:
	case <-shutdownCtx.Done():
		log.Print("timed out waiting for message broker subscriber to stop")
	}
	select {
	case <-hubDone:
	case <-shutdownCtx.Done():
		log.Print("timed out waiting for WebSocket connections to stop")
	}
	if handlers.MediaProcessor != nil {
		handlers.MediaProcessor.Close()
	}
	if handlers.Redis != nil {
		if err := handlers.Redis.Close(); err != nil {
			log.Printf("Redis shutdown: %v", err)
		}
	}
	db.Close()
	log.Print("shutdown complete")
}

func getConfigOptions() config.Options {
	cfgSource := os.Getenv(config.SourceKey)
	if cfgSource == "" {
		cfgSource = config.SourceEnv
	}
	cfgOptions := config.Options{ConfigFileSource: cfgSource}
	if cfgSource == config.SourceEnv {
		cfgOptions.ConfigFile = ".env"
	}
	return cfgOptions
}
