package slack

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	slackAPIBaseURL = "https://slack.com/api"
)

// Client represents a Slack API client
type Client struct {
	botToken      string
	signingSecret string
	httpClient    *http.Client
}

// NewClient creates a new Slack API client
func NewClient(botToken, signingSecret string) *Client {
	return &Client{
		botToken:      botToken,
		signingSecret: signingSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// UserInfo represents Slack user information
type UserInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
}

// MessageResponse represents a Slack API message response
type MessageResponse struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	Channel string `json:"channel,omitempty"`
	TS      string `json:"ts,omitempty"`
}

// ConversationHistoryResponse represents the response from conversations.history
type ConversationHistoryResponse struct {
	OK       bool `json:"ok"`
	Messages []struct {
		Type string `json:"type"`
		User string `json:"user"`
		Text string `json:"text"`
		TS   string `json:"ts"`
	} `json:"messages"`
	Error string `json:"error,omitempty"`
}

// UserInfoResponse represents the response from users.info
type UserInfoResponse struct {
	OK    bool `json:"ok"`
	User  UserInfo `json:"user"`
	Error string   `json:"error,omitempty"`
}

// VerifyRequest verifies a Slack request signature
func (c *Client) VerifyRequest(r *http.Request, body []byte) bool {
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	signature := r.Header.Get("X-Slack-Signature")

	if timestamp == "" || signature == "" {
		return false
	}

	// Check timestamp to prevent replay attacks (within 5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix()-ts > 60*5 {
		return false
	}

	// Compute expected signature
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, body)
	mac := hmac.New(sha256.New, []byte(c.signingSecret))
	mac.Write([]byte(baseString))
	expectedSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// PostMessage sends a message to a Slack channel
func (c *Client) PostMessage(channel, text string) (*MessageResponse, error) {
	payload := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}

	return c.postJSON("chat.postMessage", payload)
}

// PostMessageWithBlocks sends a message with blocks to a Slack channel
func (c *Client) PostMessageWithBlocks(channel, text string, blocks interface{}) (*MessageResponse, error) {
	payload := map[string]interface{}{
		"channel": channel,
		"text":    text,
		"blocks":  blocks,
	}

	return c.postJSON("chat.postMessage", payload)
}

// AddReaction adds an emoji reaction to a message
func (c *Client) AddReaction(channel, timestamp, emoji string) error {
	payload := map[string]interface{}{
		"channel":   channel,
		"timestamp": timestamp,
		"name":      emoji,
	}

	resp, err := c.postJSON("reactions.add", payload)
	if err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("slack API error: %s", resp.Error)
	}

	return nil
}

// GetUserInfo retrieves information about a user
func (c *Client) GetUserInfo(userID string) (*UserInfo, error) {
	url := fmt.Sprintf("%s/users.info?user=%s", slackAPIBaseURL, userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.botToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userResp UserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, err
	}

	if !userResp.OK {
		return nil, fmt.Errorf("slack API error: %s", userResp.Error)
	}

	return &userResp.User, nil
}

// GetMessageText retrieves the text of a specific message
func (c *Client) GetMessageText(channel, timestamp string) (string, error) {
	url := fmt.Sprintf("%s/conversations.history?channel=%s&latest=%s&limit=1&inclusive=true",
		slackAPIBaseURL, channel, timestamp)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.botToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var histResp ConversationHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&histResp); err != nil {
		return "", err
	}

	if !histResp.OK {
		return "", fmt.Errorf("slack API error: %s", histResp.Error)
	}

	if len(histResp.Messages) == 0 {
		return "", fmt.Errorf("message not found")
	}

	return histResp.Messages[0].Text, nil
}

// postJSON is a helper to make POST requests with JSON payload
func (c *Client) postJSON(endpoint string, payload map[string]interface{}) (*MessageResponse, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", slackAPIBaseURL, endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var msgResp MessageResponse
	if err := json.Unmarshal(body, &msgResp); err != nil {
		return nil, err
	}

	return &msgResp, nil
}
