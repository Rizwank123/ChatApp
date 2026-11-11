package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Rizwank123/chatApp/internal/dependency"
	"github.com/Rizwank123/chatApp/internal/http/swagger"
	"github.com/Rizwank123/chatApp/internal/pkg/config"
	"github.com/Rizwank123/chatApp/internal/service"
	"github.com/labstack/echo/v4"
)

func main() {
	cfgOpt := getConfigOptions()
	cfg, err := dependency.NewConfig(cfgOpt)
	if err != nil {
		log.Fatalf("failed to load configuration")
	}
	// Setup database connection
	db, err := dependency.NewDatabaseConfig(cfg)
	if err != nil {
		log.Fatalf("failed to create connection for database: %v", err)
	}

	// Initialize the dependencies
	handlers, err := dependency.NewChatAppApi(cfg, db)
	if err != nil {
		log.Fatalf("failed to create dependencies: %v", err)
	}

	api := handlers.API
	hub := handlers.Hub
	broker := handlers.Broker

	//Start the WebSocket server
	go hub.Run()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := service.DeliverBrokerMessages(ctx, broker, hub); err != nil && err != context.Canceled {
			log.Printf("message broker stopped: %v", err)
		}
	}()

	// Set up the echo server
	e := echo.New()
	e.HideBanner = true

	// Set up the middleware
	api.SetupMiddleware(e)

	// Set up the swagger documentation
	swagger.SetupSwagger(cfg, e)

	// Set up the routes
	api.SetupRoutes(e)

	// Start the server in a goroutine to handle graceful shutdown
	go func() {
		e.Logger.Info(e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.AppPort)))
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Server is shutting down...")
	cancel()

	// Create a channel to signal shutdown completion
	done := make(chan struct{})

	// Start shutdown in a goroutine
	go func() {
		if err := e.Shutdown(nil); err != nil {
			e.Logger.Fatal(err)
		}
		close(done)
	}()

	// Wait for shutdown with timeout
	select {
	case <-done:
		log.Println("Server gracefully stopped")
	case <-time.After(10 * time.Second):
		log.Println("Server shutdown timed out")
	}
}

func getConfigOptions() config.Options {
	// Load the configuration
	cfgSource := os.Getenv(config.SourceKey)
	if cfgSource == "" {
		cfgSource = config.SourceEnv
	}
	cfgOptions := config.Options{
		ConfigFileSource: cfgSource,
	}
	switch cfgSource {
	case config.SourceEnv:
		cfgOptions.ConfigFile = ".env"
	}
	return cfgOptions
}
