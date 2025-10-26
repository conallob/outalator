# Slack Bot Integration

The Outalator Slack bot allows teams to interact with the outage database directly from Slack, enabling quick note-taking and collaboration during incidents.

## Features

1. **Direct Message Commands**: Post notes or create outages using simple text commands
2. **Emoji Reactions**: Tag existing Slack messages to add them as notes to an outage

## Setup

### 1. Create a Slack App

1. Go to [Slack API Apps](https://api.slack.com/apps)
2. Click "Create New App" → "From scratch"
3. Name your app (e.g., "Outalator Bot") and select your workspace
4. Under "OAuth & Permissions", add these Bot Token Scopes:
   - `app_mentions:read` - Read mentions
   - `channels:history` - Read public channel messages
   - `channels:read` - View basic channel info
   - `chat:write` - Post messages
   - `reactions:read` - View emoji reactions
   - `reactions:write` - Add emoji reactions
   - `users:read` - View users in workspace
5. Install the app to your workspace
6. Copy the "Bot User OAuth Token" (starts with `xoxb-`)
7. Under "Basic Information", copy the "Signing Secret"

### 2. Configure Event Subscriptions

1. Under "Event Subscriptions", enable events
2. Set the Request URL to: `https://your-server.com/slack/events`
3. Subscribe to these bot events:
   - `message.channels` - Listen to messages in public channels
   - `message.groups` - Listen to messages in private channels
   - `message.im` - Listen to direct messages
   - `reaction_added` - Listen to emoji reactions
4. Save changes

### 3. Configure Outalator

Add to your `config.yaml`:

```yaml
slack:
  enabled: true
  bot_token: xoxb-your-bot-token-here
  signing_secret: your-signing-secret-here
  reaction_emoji: outage_note  # The emoji for tagging messages (without colons)
```

Or use environment variables:

```bash
export SLACK_ENABLED=true
export SLACK_BOT_TOKEN=xoxb-your-bot-token-here
export SLACK_SIGNING_SECRET=your-signing-secret-here
export SLACK_REACTION_EMOJI=outage_note
```

### 4. Create Custom Emoji (Optional)

For a better experience, create a custom emoji in your Slack workspace:

1. Go to your Slack workspace → Customize → Emoji
2. Upload an icon for `:outage_note:`
3. This emoji will be used to tag messages for adding to outages

## Usage

### Creating an Outage

Send a direct message to the bot:

```
outage API Gateway is down | Users cannot authenticate | critical
```

Format: `outage <title> | <description> | <severity>`

Severity options: `critical`, `high`, `medium`, `low`

The bot will respond with the created outage ID.

### Adding a Note to an Outage

Send a direct message to the bot:

```
note 123e4567-e89b-12d3-a456-426614174000 Restarted the API gateway service, users can now authenticate
```

Format: `note <outage_id> <content>`

The bot will add the note to the specified outage with your Slack username as the author.

### Tagging Slack Messages

1. Post a message in a Slack channel that mentions the outage ID:
   ```
   Just investigated the issue for outage 123e4567-e89b-12d3-a456-426614174000 - looks like a database connection pool exhaustion
   ```

2. React to the message with the configured emoji (default: `:outage_note:`)

3. The bot will automatically add the message content as a note to that outage and confirm with a checkmark reaction

## Architecture

The Slack integration consists of:

- **`internal/slack/bot.go`**: Event handling and command processing
- **`internal/slack/client.go`**: Slack API client for posting messages and reactions
- **`cmd/outalator/main.go`**: Main application with Slack bot initialization

The bot:
1. Verifies incoming requests using HMAC signature validation
2. Handles Slack's URL verification challenge
3. Processes events asynchronously to ensure quick responses
4. Integrates with the Outalator service layer for database operations

## Security

- All Slack requests are verified using HMAC-SHA256 signatures
- Timestamps are checked to prevent replay attacks (5-minute window)
- Only authenticated Slack users from your workspace can interact with the bot
- The bot uses OAuth tokens with minimal required scopes

## Troubleshooting

### Bot doesn't respond to messages

1. Check that the bot is invited to the channel: `/invite @Outalator`
2. Verify the bot token and signing secret are correct
3. Check server logs for authentication errors
4. Ensure the Event Subscriptions URL is accessible from Slack's servers

### Emoji reactions don't work

1. Verify the `reaction_emoji` matches the emoji name exactly (without colons)
2. Check that the bot has `reactions:read` permission
3. Ensure messages mention the outage ID in the correct format
4. Check server logs for parsing errors

### Invalid outage ID errors

1. Verify the outage ID is a valid UUID
2. Check that the outage exists in the database
3. Copy the exact ID from the bot's outage creation response

## Future Enhancements

Potential improvements:

- Interactive message components (buttons, dropdowns)
- Slash commands for quick actions
- Thread-based conversations for specific outages
- Periodic status updates posted to channels
- Integration with Slack's incident management features
