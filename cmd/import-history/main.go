package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/conall/outalator/internal/config"
	"github.com/conall/outalator/internal/domain"
	"github.com/conall/outalator/internal/notification"
	"github.com/conall/outalator/internal/notification/opsgenie"
	"github.com/conall/outalator/internal/notification/pagerduty"
	"github.com/conall/outalator/internal/storage/postgres"
	"github.com/google/uuid"
)

type ImportStats struct {
	TotalFetched  int
	NewOutages    int
	NewAlerts     int
	Skipped       int
	Errors        int
}

func main() {
	// Command-line flags
	var (
		configPath  = flag.String("config", "config.yaml", "Path to configuration file")
		service     = flag.String("service", "", "Service to import from (pagerduty or opsgenie)")
		since       = flag.String("since", "", "Start date for import (RFC3339 format, e.g., 2024-01-01T00:00:00Z)")
		until       = flag.String("until", "", "End date for import (RFC3339 format, optional)")
		teams       = flag.String("teams", "", "Comma-separated list of team IDs to filter (optional)")
		listTeams   = flag.Bool("list-teams", false, "List available teams and exit")
		dryRun      = flag.Bool("dry-run", false, "Preview what would be imported without making changes")
		batchSize   = flag.Int("batch-size", 100, "Number of incidents to fetch per API call")
	)
	flag.Parse()

	// Validate required flags
	if *service == "" {
		log.Fatal("Error: -service flag is required (pagerduty or opsgenie)")
	}

	if *service != "pagerduty" && *service != "opsgenie" {
		log.Fatal("Error: -service must be either 'pagerduty' or 'opsgenie'")
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize notification service
	var notificationService interface{}
	switch *service {
	case "pagerduty":
		if cfg.PagerDuty == nil || cfg.PagerDuty.APIKey == "" {
			log.Fatal("Error: PagerDuty API key not configured")
		}
		notificationService = pagerduty.New(pagerduty.Config{
			APIKey: cfg.PagerDuty.APIKey,
			APIURL: cfg.PagerDuty.APIURL,
		})
	case "opsgenie":
		if cfg.OpsGenie == nil || cfg.OpsGenie.APIKey == "" {
			log.Fatal("Error: OpsGenie API key not configured")
		}
		notificationService = opsgenie.New(opsgenie.Config{
			APIKey: cfg.OpsGenie.APIKey,
			APIURL: cfg.OpsGenie.APIURL,
		})
	}

	ctx := context.Background()

	// Handle list-teams flag
	if *listTeams {
		listAvailableTeams(ctx, notificationService, *service)
		return
	}

	// Validate and parse date flags
	if *since == "" {
		log.Fatal("Error: -since flag is required (RFC3339 format, e.g., 2024-01-01T00:00:00Z)")
	}

	sinceTime, err := time.Parse(time.RFC3339, *since)
	if err != nil {
		log.Fatalf("Error: Invalid -since date format: %v", err)
	}

	var untilTime time.Time
	if *until != "" {
		untilTime, err = time.Parse(time.RFC3339, *until)
		if err != nil {
			log.Fatalf("Error: Invalid -until date format: %v", err)
		}
	} else {
		untilTime = time.Now()
	}

	// Parse team filter
	var teamIDs []string
	if *teams != "" {
		teamIDs = strings.Split(*teams, ",")
		for i := range teamIDs {
			teamIDs[i] = strings.TrimSpace(teamIDs[i])
		}
	}

	// Initialize storage (only if not dry-run)
	var store *postgres.PostgresStorage
	if !*dryRun {
		store, err = postgres.New(postgres.Config{
			Host:     cfg.Database.Host,
			Port:     cfg.Database.Port,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			DBName:   cfg.Database.DBName,
			SSLMode:  cfg.Database.SSLMode,
		})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer store.Close()
	}

	// Run import
	log.Printf("Starting import from %s", *service)
	log.Printf("Date range: %s to %s", sinceTime.Format(time.RFC3339), untilTime.Format(time.RFC3339))
	if len(teamIDs) > 0 {
		log.Printf("Team filter: %v", teamIDs)
	}
	if *dryRun {
		log.Println("DRY RUN MODE - No changes will be made")
	}
	log.Println()

	stats := &ImportStats{}
	err = runImport(ctx, notificationService, store, sinceTime, untilTime, teamIDs, *batchSize, *dryRun, stats, *service)
	if err != nil {
		log.Fatalf("Import failed: %v", err)
	}

	// Print statistics
	log.Println()
	log.Println("Import completed!")
	log.Printf("Total incidents/alerts fetched: %d", stats.TotalFetched)
	log.Printf("New outages created: %d", stats.NewOutages)
	log.Printf("New alerts created: %d", stats.NewAlerts)
	log.Printf("Skipped (already exists): %d", stats.Skipped)
	if stats.Errors > 0 {
		log.Printf("Errors encountered: %d", stats.Errors)
	}
}

func listAvailableTeams(ctx context.Context, svc interface{}, serviceName string) {
	log.Printf("Fetching teams from %s...\n", serviceName)

	var teams []interface{ GetID() string; GetName() string }
	var err error

	switch serviceName {
	case "pagerduty":
		pdService := svc.(*pagerduty.Service)
		pdTeams, err := pdService.ListTeams(ctx)
		if err != nil {
			log.Fatalf("Failed to list teams: %v", err)
		}
		teams = make([]interface{ GetID() string; GetName() string }, len(pdTeams))
		for i := range pdTeams {
			teams[i] = &teamAdapter{id: pdTeams[i].ID, name: pdTeams[i].Name}
		}
	case "opsgenie":
		ogService := svc.(*opsgenie.Service)
		ogTeams, err := ogService.ListTeams(ctx)
		if err != nil {
			log.Fatalf("Failed to list teams: %v", err)
		}
		teams = make([]interface{ GetID() string; GetName() string }, len(ogTeams))
		for i := range ogTeams {
			teams[i] = &teamAdapter{id: ogTeams[i].ID, name: ogTeams[i].Name}
		}
	}

	if err != nil {
		log.Fatalf("Failed to list teams: %v", err)
	}

	log.Println("\nAvailable teams:")
	for _, team := range teams {
		fmt.Printf("  ID: %s\tName: %s\n", team.GetID(), team.GetName())
	}
}

type teamAdapter struct {
	id   string
	name string
}

func (t *teamAdapter) GetID() string   { return t.id }
func (t *teamAdapter) GetName() string { return t.name }

func runImport(
	ctx context.Context,
	svc interface{},
	store *postgres.PostgresStorage,
	since, until time.Time,
	teamIDs []string,
	batchSize int,
	dryRun bool,
	stats *ImportStats,
	serviceName string,
) error {
	offset := 0
	hasMore := true

	for hasMore {
		var alerts []*notification.Alert
		var err error

		// Fetch batch based on service type
		switch serviceName {
		case "pagerduty":
			pdService := svc.(*pagerduty.Service)
			opts := pagerduty.HistoricalFetchOptions{
				Since:   since,
				Until:   until,
				TeamIDs: teamIDs,
				Limit:   batchSize,
				Offset:  offset,
			}
			alerts, hasMore, err = pdService.FetchHistoricalIncidents(ctx, opts)
		case "opsgenie":
			ogService := svc.(*opsgenie.Service)
			opts := opsgenie.HistoricalFetchOptions{
				Since:   since,
				Until:   until,
				TeamIDs: teamIDs,
				Limit:   batchSize,
				Offset:  offset,
			}
			alerts, hasMore, err = ogService.FetchHistoricalAlerts(ctx, opts)
		}

		if err != nil {
			return fmt.Errorf("failed to fetch alerts at offset %d: %w", offset, err)
		}

		if len(alerts) == 0 {
			break
		}

		stats.TotalFetched += len(alerts)
		log.Printf("Fetched %d incidents/alerts (offset: %d)", len(alerts), offset)

		// Process each alert
		for _, alert := range alerts {
			if err := processAlert(ctx, store, alert, dryRun, stats); err != nil {
				log.Printf("Error processing alert %s: %v", alert.ExternalID, err)
				stats.Errors++
			}
		}

		offset += batchSize

		// Small delay to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func processAlert(
	ctx context.Context,
	store *postgres.PostgresStorage,
	alert *notification.Alert,
	dryRun bool,
	stats *ImportStats,
) error {
	if dryRun {
		log.Printf("  [DRY RUN] Would import: %s - %s (Team: %s, Date: %s)",
			alert.ExternalID, alert.Title, alert.TeamName, alert.TriggeredAt.Format(time.RFC3339))
		stats.NewOutages++
		stats.NewAlerts++
		return nil
	}

	// Check if alert already exists
	existing, err := store.GetAlertByExternalID(ctx, alert.ExternalID, alert.Source)
	if err != nil && err.Error() != "alert not found" {
		return fmt.Errorf("failed to check existing alert: %w", err)
	}

	if existing != nil {
		log.Printf("  Skipping %s - already exists", alert.ExternalID)
		stats.Skipped++
		return nil
	}

	// Create a new outage for this alert
	outageID := uuid.New()
	status := "resolved"
	if alert.ResolvedAt == nil {
		status = "open"
	}

	outage := &domain.Outage{
		ID:          outageID,
		Title:       alert.Title,
		Description: alert.Description,
		Status:      status,
		Severity:    alert.Severity,
		CreatedAt:   alert.TriggeredAt,
		UpdatedAt:   alert.TriggeredAt,
		ResolvedAt:  alert.ResolvedAt,
	}

	if err := store.CreateOutage(ctx, outage); err != nil {
		return fmt.Errorf("failed to create outage: %w", err)
	}
	stats.NewOutages++

	// Create the alert and link it to the outage
	domainAlert := &domain.Alert{
		ID:             uuid.New(),
		OutageID:       outageID,
		ExternalID:     alert.ExternalID,
		Source:         alert.Source,
		TeamName:       alert.TeamName,
		Title:          alert.Title,
		Description:    alert.Description,
		Severity:       alert.Severity,
		TriggeredAt:    alert.TriggeredAt,
		AcknowledgedAt: alert.AcknowledgedAt,
		ResolvedAt:     alert.ResolvedAt,
		CreatedAt:      time.Now(),
	}

	if err := store.CreateAlert(ctx, domainAlert); err != nil {
		// Try to clean up the outage if alert creation fails
		_ = store.DeleteOutage(ctx, outageID)
		return fmt.Errorf("failed to create alert: %w", err)
	}
	stats.NewAlerts++

	log.Printf("  Imported: %s - %s (Team: %s)", alert.ExternalID, alert.Title, alert.TeamName)

	return nil
}
