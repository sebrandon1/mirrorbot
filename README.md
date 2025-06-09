# MirrorBot

A Slack bot for querying OpenShift release status and details.

## Prerequisites

- Go 1.24+ (for local development)
- Docker (for containerized runs)
- [Slack App](https://api.slack.com/apps) with:
  - Bot Token (`SLACK_BOT_TOKEN`)
  - App Token (`SLACK_APP_TOKEN`) with Socket Mode enabled
- (Optional) [golangci-lint](https://golangci-lint.run/usage/install/) for linting

## How to Run

### 1. Clone the repository

```bash
git clone https://github.com/sebrandon1/mirrorbot.git
cd mirrorbot
```

### 2. Set up your Slack tokens

Export your Slack tokens as environment variables:

```bash
export SLACK_BOT_TOKEN='xoxb-...'
export SLACK_APP_TOKEN='xapp-...'
```

### 3. Run locally (with Go)

```bash
make build
./mirrorbot
```

### 4. Run in Docker

Build and run the image:

```bash
make docker-build
make run-image
```

> **Note:** The `run-image` target checks for your Slack tokens and passes them into the container.

### 5. (Optional) Lint and Test

```bash
make vet
make lint
make test
```

## Usage

Invite the bot to a channel and mention it with an OpenShift version, e.g.:

```
@mirrorbot 4.20
```

The bot will respond with the latest release info for that version.

---

For more details, see the code and comments in [main.go](main.go).
