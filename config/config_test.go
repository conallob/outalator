package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want 0.0.0.0", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %d, want 5432", cfg.Database.Port)
	}
	if cfg.GRPC.Enabled {
		t.Error("GRPC.Enabled should default to false")
	}
}

func TestLoadFromYAML(t *testing.T) {
	yaml := `
server:
  host: "127.0.0.1"
  port: 9000
database:
  host: "dbhost"
  port: 5432
  user: "testuser"
  password: "testpass"
  dbname: "testdb"
  sslmode: "require"
grpc:
  enabled: true
  port: 9090
`
	path := writeConfig(t, yaml)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %q, want 127.0.0.1", cfg.Server.Host)
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("Server.Port = %d, want 9000", cfg.Server.Port)
	}
	if cfg.Database.Host != "dbhost" {
		t.Errorf("Database.Host = %q, want dbhost", cfg.Database.Host)
	}
	if cfg.Database.User != "testuser" {
		t.Errorf("Database.User = %q, want testuser", cfg.Database.User)
	}
	if !cfg.GRPC.Enabled {
		t.Error("GRPC.Enabled should be true")
	}
	if cfg.GRPC.Port != 9090 {
		t.Errorf("GRPC.Port = %d, want 9090", cfg.GRPC.Port)
	}
}

func TestLoadPagerDutyAndOpsGenie(t *testing.T) {
	yaml := `
server:
  port: 8080
pagerduty:
  api_key: "pd-key"
  api_url: "https://pd.example.com"
opsgenie:
  api_key: "og-key"
`
	path := writeConfig(t, yaml)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.PagerDuty == nil {
		t.Fatal("PagerDuty config is nil")
	}
	if cfg.PagerDuty.APIKey != "pd-key" {
		t.Errorf("PagerDuty.APIKey = %q, want pd-key", cfg.PagerDuty.APIKey)
	}
	if cfg.OpsGenie == nil {
		t.Fatal("OpsGenie config is nil")
	}
	if cfg.OpsGenie.APIKey != "og-key" {
		t.Errorf("OpsGenie.APIKey = %q, want og-key", cfg.OpsGenie.APIKey)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	yaml := `
server:
  host: "0.0.0.0"
  port: 8080
database:
  host: "localhost"
  port: 5432
`
	path := writeConfig(t, yaml)

	// Set env vars
	t.Setenv("SERVER_HOST", "10.0.0.1")
	t.Setenv("SERVER_PORT", "9999")
	t.Setenv("DB_HOST", "remote-db")
	t.Setenv("DB_USER", "envuser")
	t.Setenv("DB_PASSWORD", "envpass")
	t.Setenv("DB_NAME", "envdb")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.Host != "10.0.0.1" {
		t.Errorf("Server.Host = %q, want 10.0.0.1", cfg.Server.Host)
	}
	if cfg.Server.Port != 9999 {
		t.Errorf("Server.Port = %d, want 9999", cfg.Server.Port)
	}
	if cfg.Database.Host != "remote-db" {
		t.Errorf("Database.Host = %q, want remote-db", cfg.Database.Host)
	}
	if cfg.Database.User != "envuser" {
		t.Errorf("Database.User = %q, want envuser", cfg.Database.User)
	}
}

func TestLoadEnvPagerDutyKey(t *testing.T) {
	yaml := `server: {port: 8080}`
	path := writeConfig(t, yaml)

	t.Setenv("PAGERDUTY_API_KEY", "env-pd-key")
	t.Setenv("OPSGENIE_API_KEY", "env-og-key")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.PagerDuty == nil || cfg.PagerDuty.APIKey != "env-pd-key" {
		t.Errorf("PagerDuty.APIKey = %v, want env-pd-key", cfg.PagerDuty)
	}
	if cfg.OpsGenie == nil || cfg.OpsGenie.APIKey != "env-og-key" {
		t.Errorf("OpsGenie.APIKey = %v, want env-og-key", cfg.OpsGenie)
	}
}

func TestLoadAuthConfig(t *testing.T) {
	yaml := `
server: {port: 8080}
auth:
  enabled: true
  issuer: "https://issuer.example.com"
  client_id: "my-client"
  client_secret: "my-secret"
  redirect_url: "https://app.example.com/callback"
`
	path := writeConfig(t, yaml)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Auth == nil {
		t.Fatal("Auth config is nil")
	}
	if !cfg.Auth.Enabled {
		t.Error("Auth.Enabled should be true")
	}
	if cfg.Auth.Issuer != "https://issuer.example.com" {
		t.Errorf("Auth.Issuer = %q", cfg.Auth.Issuer)
	}
	if cfg.Auth.ClientID != "my-client" {
		t.Errorf("Auth.ClientID = %q", cfg.Auth.ClientID)
	}
}

func TestLoadAuthEnvOverrides(t *testing.T) {
	yaml := `server: {port: 8080}`
	path := writeConfig(t, yaml)

	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_ISSUER", "https://env-issuer.com")
	t.Setenv("AUTH_CLIENT_ID", "env-client")
	t.Setenv("AUTH_CLIENT_SECRET", "env-secret")
	t.Setenv("AUTH_REDIRECT_URL", "https://env-app.com/cb")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Auth == nil {
		t.Fatal("Auth config is nil")
	}
	if !cfg.Auth.Enabled {
		t.Error("Auth.Enabled should be true")
	}
	if cfg.Auth.Issuer != "https://env-issuer.com" {
		t.Errorf("Auth.Issuer = %q, want https://env-issuer.com", cfg.Auth.Issuer)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/does/not/exist.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	path := writeConfig(t, "not: [valid: yaml{{{")
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadSlackConfig(t *testing.T) {
	yaml := `
server: {port: 8080}
slack:
  enabled: true
  bot_token: "xoxb-token"
  signing_secret: "slack-secret"
  reaction_emoji: "outage_note"
`
	path := writeConfig(t, yaml)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Slack == nil {
		t.Fatal("Slack config is nil")
	}
	if !cfg.Slack.Enabled {
		t.Error("Slack.Enabled should be true")
	}
	if cfg.Slack.BotToken != "xoxb-token" {
		t.Errorf("Slack.BotToken = %q", cfg.Slack.BotToken)
	}
	if cfg.Slack.ReactionEmoji != "outage_note" {
		t.Errorf("Slack.ReactionEmoji = %q", cfg.Slack.ReactionEmoji)
	}
}

func TestGRPCEnvOverride(t *testing.T) {
	yaml := `server: {port: 8080}`
	path := writeConfig(t, yaml)

	t.Setenv("GRPC_ENABLED", "true")
	t.Setenv("GRPC_HOST", "grpc-host")
	t.Setenv("GRPC_PORT", "9191")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.GRPC.Enabled {
		t.Error("GRPC.Enabled should be true")
	}
	if cfg.GRPC.Host != "grpc-host" {
		t.Errorf("GRPC.Host = %q, want grpc-host", cfg.GRPC.Host)
	}
	if cfg.GRPC.Port != 9191 {
		t.Errorf("GRPC.Port = %d, want 9191", cfg.GRPC.Port)
	}
}
