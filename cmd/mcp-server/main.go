package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/conall/outalator/internal/config"
	"github.com/conall/outalator/internal/mcp"
	"github.com/conall/outalator/internal/notification/opsgenie"
	"github.com/conall/outalator/internal/notification/pagerduty"
	"github.com/conall/outalator/internal/service"
	"github.com/conall/outalator/internal/storage/postgres"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	db, err := postgres.NewStorage(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize service
	svc := service.New(db)

	// Register notification services
	if cfg.PagerDuty != nil && cfg.PagerDuty.APIKey != "" {
		pdConfig := pagerduty.Config{
			APIKey: cfg.PagerDuty.APIKey,
			APIURL: cfg.PagerDuty.APIURL,
		}
		pdSvc := pagerduty.New(pdConfig)
		svc.RegisterNotificationService(pdSvc)
		log.Println("Registered PagerDuty notification service")
	}

	if cfg.OpsGenie != nil && cfg.OpsGenie.APIKey != "" {
		ogConfig := opsgenie.Config{
			APIKey: cfg.OpsGenie.APIKey,
			APIURL: cfg.OpsGenie.APIURL,
		}
		ogSvc := opsgenie.New(ogConfig)
		svc.RegisterNotificationService(ogSvc)
		log.Println("Registered OpsGenie notification service")
	}

	// Create MCP server
	mcpServer := mcp.NewServer(svc)

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, shutting down...")
		cancel()
	}()

	// Serve MCP over stdio
	log.Println("Starting MCP server on stdio...")
	if err := mcpServer.ServeStdio(ctx, os.Stdin, os.Stdout); err != nil && err != context.Canceled {
		log.Fatalf("MCP server error: %v", err)
	}

	fmt.Fprintln(os.Stderr, "MCP server stopped")
}
