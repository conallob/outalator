package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/conall/outalator/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Bot represents a Slack bot instance
type Bot struct {
	service       *service.Service
	client        *Client
	reactionEmoji string // The emoji used to tag messages for note creation
}

// Config holds Slack bot configuration
type Config struct {
	SigningSecret string
	BotToken      string
	ReactionEmoji string // e.g., "outage_note" for :outage_note:
}

// NewBot creates a new Slack bot instance
func NewBot(svc *service.Service, cfg Config) *Bot {
	client := NewClient(cfg.BotToken, cfg.SigningSecret)
	return &Bot{
		service:       svc,
		client:        client,
		reactionEmoji: cfg.ReactionEmoji,
	}
}

// SlackEvent represents a Slack event
type SlackEvent struct {
	Type      string          `json:"type"`
	Challenge string          `json:"challenge,omitempty"` // For URL verification
	Event     json.RawMessage `json:"event,omitempty"`
}

// MessageEvent represents a Slack message event
type MessageEvent struct {
	Type    string `json:"type"`
	User    string `json:"user"`
	Text    string `json:"text"`
	Channel string `json:"channel"`
	TS      string `json:"ts"` // Timestamp, used as message ID
}

// ReactionAddedEvent represents a Slack reaction_added event
type ReactionAddedEvent struct {
	Type     string `json:"type"`
	User     string `json:"user"`
	Reaction string `json:"reaction"`
	Item     struct {
		Type    string `json:"type"`
		Channel string `json:"channel"`
		TS      string `json:"ts"`
	} `json:"item"`
}

// HandleEvent processes incoming Slack events
func (b *Bot) HandleEvent(w http.ResponseWriter, r *http.Request) {
	// Read body for verification
	body := new(bytes.Buffer)
	body.ReadFrom(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body.Bytes()))

	// Verify request signature
	if !b.client.VerifyRequest(r, body.Bytes()) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Restore body for decoding
	r.Body = io.NopCloser(bytes.NewBuffer(body.Bytes()))

	var event SlackEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Printf("Error decoding event: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Handle URL verification challenge
	if event.Type == "url_verification" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"challenge": event.Challenge,
		})
		return
	}

	// Process event asynchronously
	go b.processEvent(event)

	w.WriteHeader(http.StatusOK)
}

// processEvent handles different types of Slack events
func (b *Bot) processEvent(event SlackEvent) {
	var eventType struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(event.Event, &eventType); err != nil {
		log.Printf("Error parsing event type: %v", err)
		return
	}

	ctx := context.Background()

	switch eventType.Type {
	case "message":
		b.handleMessage(ctx, event.Event)
	case "reaction_added":
		b.handleReactionAdded(ctx, event.Event)
	}
}

// handleMessage processes direct message commands
func (b *Bot) handleMessage(ctx context.Context, eventData json.RawMessage) {
	var msg MessageEvent
	if err := json.Unmarshal(eventData, &msg); err != nil {
		log.Printf("Error parsing message event: %v", err)
		return
	}

	// Ignore bot messages
	if msg.User == "" {
		return
	}

	// Parse outage note command
	// Format: "note <outage_id> <content>"
	if strings.HasPrefix(msg.Text, "note ") {
		b.handleNoteCommand(ctx, msg)
		return
	}

	// Parse create outage command
	// Format: "outage <title> | <description> | <severity>"
	if strings.HasPrefix(msg.Text, "outage ") {
		b.handleOutageCommand(ctx, msg)
		return
	}
}

// handleNoteCommand processes the "note" command
func (b *Bot) handleNoteCommand(ctx context.Context, msg MessageEvent) {
	// Parse: "note <outage_id> <content>"
	pattern := regexp.MustCompile(`^note\s+([a-fA-F0-9-]+)\s+(.+)$`)
	matches := pattern.FindStringSubmatch(msg.Text)

	if len(matches) != 3 {
		b.sendMessage(msg.Channel, "Invalid format. Use: `note <outage_id> <content>`")
		return
	}

	outageID, err := uuid.Parse(matches[1])
	if err != nil {
		b.sendMessage(msg.Channel, fmt.Sprintf("Invalid outage ID: %v", err))
		return
	}

	content := matches[2]

	// Get user info for author
	author := b.getUserName(msg.User)

	req := struct {
		Content string `json:"content"`
		Format  string `json:"format"`
		Author  string `json:"author"`
	}{
		Content: content,
		Format:  "plaintext",
		Author:  author,
	}

	note, err := b.service.AddNote(ctx, outageID, req)
	if err != nil {
		b.sendMessage(msg.Channel, fmt.Sprintf("Error adding note: %v", err))
		return
	}

	b.sendMessage(msg.Channel, fmt.Sprintf("✅ Added note to outage %s (Note ID: %s)", outageID, note.ID))
}

// handleOutageCommand processes the "outage" command
func (b *Bot) handleOutageCommand(ctx context.Context, msg MessageEvent) {
	// Parse: "outage <title> | <description> | <severity>"
	parts := strings.Split(strings.TrimPrefix(msg.Text, "outage "), "|")
	if len(parts) != 3 {
		b.sendMessage(msg.Channel, "Invalid format. Use: `outage <title> | <description> | <severity>`")
		return
	}

	title := strings.TrimSpace(parts[0])
	description := strings.TrimSpace(parts[1])
	severity := strings.TrimSpace(parts[2])

	// Validate severity
	validSeverities := map[string]bool{
		"critical": true,
		"high":     true,
		"medium":   true,
		"low":      true,
	}

	if !validSeverities[severity] {
		b.sendMessage(msg.Channel, "Invalid severity. Use: critical, high, medium, or low")
		return
	}

	req := struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Severity    string   `json:"severity"`
		AlertIDs    []string `json:"alert_ids"`
		Tags        []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"tags,omitempty"`
	}{
		Title:       title,
		Description: description,
		Severity:    severity,
		Tags: []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{
			{Key: "slack_channel", Value: msg.Channel},
			{Key: "slack_user", Value: msg.User},
		},
	}

	outage, err := b.service.CreateOutage(ctx, req)
	if err != nil {
		b.sendMessage(msg.Channel, fmt.Sprintf("Error creating outage: %v", err))
		return
	}

	b.sendMessage(msg.Channel, fmt.Sprintf("✅ Created outage: %s (ID: %s, Severity: %s)", outage.Title, outage.ID, outage.Severity))
}

// handleReactionAdded processes emoji reactions
func (b *Bot) handleReactionAdded(ctx context.Context, eventData json.RawMessage) {
	var reaction ReactionAddedEvent
	if err := json.Unmarshal(eventData, &reaction); err != nil {
		log.Printf("Error parsing reaction event: %v", err)
		return
	}

	// Only process if it's the configured emoji
	if reaction.Reaction != b.reactionEmoji {
		return
	}

	// Get the original message
	messageText, err := b.getMessageText(reaction.Item.Channel, reaction.Item.TS)
	if err != nil {
		log.Printf("Error getting message text: %v", err)
		return
	}

	// Try to extract outage ID from the message
	// Expected format in thread or message: mentions outage ID like "outage abc-123-def"
	pattern := regexp.MustCompile(`(?i)outage[:\s]+([a-fA-F0-9-]{36})`)
	matches := pattern.FindStringSubmatch(messageText)

	if len(matches) < 2 {
		// If no outage ID found in message, send a helpful message
		b.sendMessage(reaction.Item.Channel, fmt.Sprintf("<@%s> Please include the outage ID in your message. Format: `outage <outage_id>`", reaction.User))
		return
	}

	outageID, err := uuid.Parse(matches[1])
	if err != nil {
		log.Printf("Invalid outage ID in message: %v", err)
		return
	}

	// Get user info for author
	author := b.getUserName(reaction.User)

	// Add the message as a note
	req := struct {
		Content string `json:"content"`
		Format  string `json:"format"`
		Author  string `json:"author"`
	}{
		Content: messageText,
		Format:  "plaintext",
		Author:  author,
	}

	note, err := b.service.AddNote(ctx, outageID, req)
	if err != nil {
		log.Printf("Error adding note from reaction: %v", err)
		b.sendMessage(reaction.Item.Channel, fmt.Sprintf("Error adding note: %v", err))
		return
	}

	// React to confirm
	b.addReaction(reaction.Item.Channel, reaction.Item.TS, "white_check_mark")
	log.Printf("Added note %s to outage %s from reaction by %s", note.ID, outageID, author)
}

// Utility methods for Slack API interactions

func (b *Bot) sendMessage(channel, text string) error {
	resp, err := b.client.PostMessage(channel, text)
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("slack error: %s", resp.Error)
	}
	return nil
}

func (b *Bot) getUserName(userID string) string {
	user, err := b.client.GetUserInfo(userID)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		return userID
	}
	if user.RealName != "" {
		return user.RealName
	}
	return user.Name
}

func (b *Bot) getMessageText(channel, timestamp string) (string, error) {
	return b.client.GetMessageText(channel, timestamp)
}

func (b *Bot) addReaction(channel, timestamp, emoji string) error {
	return b.client.AddReaction(channel, timestamp, emoji)
}

// RegisterHandlers registers HTTP handlers for the Slack bot
// Works with any router that has a HandleFunc method
func (b *Bot) RegisterHandlers(router interface {
	HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route
}) {
	router.HandleFunc("/slack/events", b.HandleEvent)
}
