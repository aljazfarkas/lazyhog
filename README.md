# lazyhog 🦔

A blazing-fast Terminal User Interface for PostHog, built with Go and the Charm stack.

**Like lazygit, but for PostHog** - Navigate your PostHog data with a beautiful sidebar interface.

## Features

### 🎯 Unified Interface
Launch `lazyhog` to get a **lazygit-style interface** with sidebar navigation. Switch between views with arrow keys, focus panels with Tab, and never leave your terminal.

```
┌─────────────┬──────────────────────────────────────┐
│ lazyhog 🦔  │  📡 Live Events                      │
│             │                                      │
│ ▶ 📡 Live   │  2026-01-16 22:15:32                 │
│   🚩 Flags  │  $pageview /dashboard                │
│   🔍 Query  │  user@example.com                    │
│             │                                      │
│             │  2026-01-16 22:15:28                 │
│             │  button_clicked upgrade_cta          │
│             │  user@example.com                    │
└─────────────┴──────────────────────────────────────┘
```

### 📡 Live Events Stream
Stream events in real-time as they happen in your PostHog instance. Navigate with arrow keys, press Enter to expand JSON details, and see events update every 2 seconds.

### 🚩 Feature Flags Manager
View and toggle feature flags instantly. Use fuzzy search to find flags, Space to toggle them on/off, and see real-time status updates.

### 👤 Person Lookup
Look up any person by their distinct_id. View their properties in a scrollable panel alongside their recent events in a two-column layout.

### 🔍 HogQL Console
Execute HogQL queries directly from your terminal. View results in a dynamic table, export to CSV with Ctrl+S, and navigate query history with arrow keys.

## Installation

### From Source
```bash
git clone https://github.com/aljazfarkas/lazyhog
cd lazyhog
go build -o lazyhog cmd/ph/*.go
sudo mv lazyhog /usr/local/bin/
```

### Using Go Install
```bash
go install github.com/aljazfarkas/lazyhog/cmd/ph@latest
# Note: Binary will be named 'ph', create alias: alias lazyhog=ph
```

## Quick Start

### 1. Authenticate

```bash
# Interactive mode (recommended)
lazyhog login

# Non-interactive mode
lazyhog login --api-key=phx_xxx --instance-url=https://app.posthog.com
```

**Important:** You need a **Personal API Key** (starts with `phx_`) from PostHog → Settings → **Personal API Keys**.
Project API Keys (`phc_`) are for sending events and won't work with lazyhog.

For self-hosted instances, provide your custom URL.

Configuration is saved to `~/.config/ph-tui.yaml` with restricted permissions.

### 2. Launch the unified interface

```bash
# Start the lazygit-style interface (recommended)
lazyhog

# Or use individual commands:
lazyhog live     # Stream live events only
lazyhog flags    # Manage feature flags only
lazyhog person user@example.com  # Look up a person
lazyhog query    # Open HogQL console only
```

### 3. Navigate the unified interface

- **↑/↓** or **j/k** - Switch between views in sidebar
- **Tab** or **→** - Focus the main panel
- **←** or **Esc** - Return to sidebar
- **q** - Quit from sidebar, or return to sidebar from panel
- **Ctrl+C** - Always quits

## Commands

### `lazyhog login`
Configure your PostHog API credentials.

**Options:**
- `--api-key` - PostHog Personal API key (must start with phx_)
- `--instance-url` - PostHog instance URL (default: https://app.posthog.com)

**Examples:**
```bash
# Interactive mode
lazyhog login

# With flags
lazyhog login --api-key=phx_xxx
```

### `lazyhog live`
Stream events in real-time.

**Keyboard shortcuts:**
- `↑/↓` or `j/k` - Navigate events
- `Enter` or `Space` - Expand/collapse event details
- `r` - Refresh
- `q` or `Ctrl+C` - Quit

### `lazyhog flags`
View and manage feature flags.

**Keyboard shortcuts:**
- `↑/↓` or `j/k` - Navigate flags
- `Space` - Toggle flag on/off
- `/` - Enter search mode
- `r` - Refresh
- `Esc` - Exit search mode
- `q` or `Ctrl+C` - Quit

### `lazyhog person [distinct_id]`
Look up a person and their recent activity.

**Keyboard shortcuts:**
- `Tab` - Switch between properties and events columns
- `↑/↓` or `j/k` - Scroll within active column
- `r` - Refresh
- `q` or `Ctrl+C` - Quit

**Example:**
```bash
lazyhog person user@example.com
lazyhog person 12345
```

### `lazyhog query`
Open the HogQL query console.

**Keyboard shortcuts:**
- `Ctrl+Enter` - Execute query
- `↑/↓` - Navigate query history (in input mode)
- `Arrow keys` - Scroll table (in result mode)
- `Ctrl+S` - Export results to CSV
- `Esc` - Return to query input
- `Ctrl+C` - Quit

**Example queries:**
```sql
-- Top events in the last day
SELECT event, count() FROM events
WHERE timestamp > now() - INTERVAL 1 DAY
GROUP BY event
ORDER BY count() DESC
LIMIT 10

-- Most active users
SELECT distinct_id, count() as event_count
FROM events
GROUP BY distinct_id
ORDER BY event_count DESC
LIMIT 10

-- Pageview paths
SELECT properties.$current_url, count()
FROM events
WHERE event = '$pageview'
GROUP BY properties.$current_url
ORDER BY count() DESC
```

## Configuration

Configuration is stored in `~/.config/ph-tui.yaml`:

```yaml
project_api_key: phx_xxxxx  # Must be Personal API key (phx_), not Project API key (phc_)
instance_url: https://app.posthog.com
poll_interval: 2  # seconds
```

## Development

### Prerequisites
- Go 1.21 or later
- Access to a PostHog instance

### Building from source
```bash
git clone https://github.com/aljazfarkas/lazyhog
cd lazyhog
go mod download
go build -o lazyhog cmd/ph/*.go
./lazyhog  # Launch the unified interface
```

### Running tests
```bash
go test ./...
```

### Project structure
```
lazyhog/
├── cmd/ph/              # CLI commands
├── internal/
│   ├── client/          # PostHog API client
│   ├── config/          # Configuration management
│   ├── ui/              # Bubbletea UI components
│   │   ├── live/        # Live events view
│   │   ├── flags/       # Feature flags view
│   │   ├── person/      # Person lookup view
│   │   ├── query/       # HogQL console
│   │   ├── components/  # Reusable UI components
│   │   └── styles/      # Styling system
│   └── utils/           # Utility functions
```

### Built with
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## Troubleshooting

### "config file not found" error
Run `lazyhog login` to set up authentication first.

### API connection errors
- Verify your API key is correct in `~/.config/ph-tui.yaml`
- **Authentication error (401)**: You're likely using a Project API key (`phc_`) instead of a Personal API key (`phx_`). Run `lazyhog login` and use a Personal API key from PostHog → Settings → Personal API Keys
- Check your internet connection
- For self-hosted instances, verify the instance URL is accessible

### Events not showing
- Check that events are being sent to your PostHog instance
- Try refreshing with `r`
- Verify your API key has the correct permissions

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT

## Acknowledgments

- PostHog team for the amazing product and API
- Charm team for the incredible TUI libraries
