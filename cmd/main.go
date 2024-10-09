package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/chatApp/internal/dependency"
	"github.com/chatApp/internal/http/swagger"
	"github.com/chatApp/internal/pkg/config"
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
	err = db.Ping(context.Background())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	// Defer closing the database connection
	defer db.Close()

	// Initialize the dependencies
	api, err := dependency.NewChatAppApi(cfg, db)

	if err != nil {
		log.Fatalf("failed to create dependencies: %v", err)
	}
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	log.Println("Server gracefully stopped")
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
