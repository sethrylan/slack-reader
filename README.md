# slack-reader

A read-only Slack CLI tool written in Go. Uses cookie-based authentication via [rneatherway/slack](https://github.com/rneatherway/slack) to access the Slack API without requiring a bot token.

## Install

```sh
go install github.com/sethrylan/slack-reader@latest
```

## Authentication

This tool uses cookie-based authentication extracted from the local Slack Desktop app. Import credentials first:

```sh
slack-reader auth creds --workspace myteam
```

Verify authentication:

```sh
slack-reader auth whoami --workspace myteam
```

## Usage

All commands require `--workspace <domain>` where `<domain>` is the Slack team domain (the `<domain>` in `<domain>.slack.com`).

The default output format is JSON. Use `--output markdown` on `message list` for a human-readable format.

### Timestamps

The `--ts` flag accepts Slack timestamps with or without the dot separator:

```
1770165109.628379   (canonical format)
1770165109628379    (concatenated digits)
```

Both forms are automatically normalized to the canonical `seconds.microseconds` format used by the Slack API.

### Messages

```sh
# Get a single message
slack-reader message get "#general" --workspace myteam --ts "1770165109.628379"

# List recent channel messages
slack-reader message list "#general" --workspace myteam

# List recent channel messages with a limit
slack-reader message list "#general" --workspace myteam --limit 50

# List all messages in a thread
slack-reader message list "#general" --workspace myteam --ts "1770165109.628379"

# Timestamps without the dot also work
slack-reader message list "#general" --workspace myteam --ts "1770165109628379"

# Output as markdown instead of JSON
slack-reader message list "#general" --workspace myteam --ts "1770165109.628379" --output markdown

# Channel IDs also work
slack-reader message get C01ABCDEF --workspace myteam --ts "1770165109.628379"
```

### Channels

```sh
# List conversations for current user
slack-reader channel list --workspace myteam

# List conversations for a specific user
slack-reader channel list --workspace myteam --user "@alice" --limit 50

# List all workspace conversations
slack-reader channel list --workspace myteam --all --limit 100
```

### Command Reference

| Command | Description |
|---------|-------------|
| `auth whoami` | Show current auth info (calls `auth.test`) |
| `auth creds` | Import credentials from Slack Desktop |
| `message get <channel> --ts <ts>` | Fetch a single message |
| `message list <channel>` | List recent channel messages |
| `message list <channel> --ts <ts>` | List all messages in a thread |
| `channel list` | List conversations for current user |
| `channel list --user "@handle"` | List conversations for a specific user |
| `channel list --all` | List all workspace conversations |

### Global Flags

| Flag | Description |
|------|-------------|
| `--workspace <domain>` | Slack team domain (required) |

### Command Flags

| Flag | Commands | Description | Default |
|------|----------|-------------|---------|
| `--ts <timestamp>` | `message get` | Message timestamp (required); with or without dot | - |
| `--ts <timestamp>` | `message list` | Thread root timestamp (with or without dot); omit to list recent channel messages | - |
| `--output <format>` | `message list` | Output format: `json` or `markdown` | `json` |
| `--user <handle>` | `channel list` | List channels for a specific user | current user |
| `--all` | `channel list` | List all workspace conversations | `false` |
| `--limit <n>` | `channel list` | Maximum results | `100` |
| `--limit <n>` | `message list` | Maximum results (`0` = unlimited) | `0` |

## License

[MIT](LICENSE)
