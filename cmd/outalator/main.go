package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/conall/outalator/internal/api"
	"github.com/conall/outalator/internal/config"
	grpcserver "github.com/conall/outalator/internal/grpc"
	"github.com/conall/outalator/internal/notification/opsgenie"
	"github.com/conall/outalator/internal/notification/pagerduty"
	"github.com/conall/outalator/internal/service"
	"github.com/conall/outalator/internal/slack"
	"github.com/conall/outalator/internal/storage/postgres"
	"github.com/gorilla/mux"
)

func main() {
	// CLI flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	slackEnabled := flag.Bool("slack-enabled", false, "Enable Slack bot integration")
	slackBotToken := flag.String("slack-bot-token", "", "Slack bot OAuth token")
	slackSigningSecret := flag.String("slack-signing-secret", "", "Slack signing secret for request verification")
	slackReactionEmoji := flag.String("slack-reaction-emoji", "", "Slack emoji name (without colons) for tagging messages as outage notes")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Apply CLI flag overrides for Slack
	if *slackEnabled {
		if cfg.Slack == nil {
			cfg.Slack = &config.SlackConfig{}
		}
		cfg.Slack.Enabled = true
	}
	if *slackBotToken != "" {
		if cfg.Slack == nil {
			cfg.Slack = &config.SlackConfig{}
		}
		cfg.Slack.BotToken = *slackBotToken
	}
	if *slackSigningSecret != "" {
		if cfg.Slack == nil {
			cfg.Slack = &config.SlackConfig{}
		}
		cfg.Slack.SigningSecret = *slackSigningSecret
	}
	if *slackReactionEmoji != "" {
		if cfg.Slack == nil {
			cfg.Slack = &config.SlackConfig{}
		}
		cfg.Slack.ReactionEmoji = *slackReactionEmoji
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

	// Set up HTTP router
	router := mux.NewRouter()

	// Register API handlers
	apiHandler := api.NewHandler(svc)
	apiHandler.RegisterRoutes(router)

	// Register Slack bot if enabled
	if cfg.Slack != nil && cfg.Slack.Enabled {
		if cfg.Slack.BotToken == "" || cfg.Slack.SigningSecret == "" {
			log.Fatal("Slack bot is enabled but bot_token or signing_secret is missing")
		}

		slackConfig := slack.Config{
			BotToken:      cfg.Slack.BotToken,
			SigningSecret: cfg.Slack.SigningSecret,
			ReactionEmoji: cfg.Slack.ReactionEmoji,
		}

		if slackConfig.ReactionEmoji == "" {
			slackConfig.ReactionEmoji = "outage_note" // Default emoji
		}

		slackBot := slack.NewBot(svc, slackConfig)
		slackBot.RegisterHandlers(router)
		log.Printf("Slack bot enabled with reaction emoji: %s", slackConfig.ReactionEmoji)
	}

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start gRPC server if enabled
	var grpcSrv *grpcserver.Server
	if cfg.GRPC.Enabled {
		grpcSrv = grpcserver.NewServer(svc)
		grpcAddr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)

		go func() {
			log.Printf("Starting gRPC server on %s", grpcAddr)
			if err := grpcSrv.Start(grpcAddr); err != nil {
				log.Fatalf("Failed to start gRPC server: %v", err)
			}
		}()
	}

	// Start HTTP server
	go func() {
		log.Printf("Starting HTTP server on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if grpcSrv != nil {
		grpcSrv.Stop()
	}

	log.Println("Servers stopped")
}
