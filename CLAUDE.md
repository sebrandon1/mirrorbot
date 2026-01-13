# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Slack bot for querying OpenShift release status and details. It connects to OCP mirrors and provides release information directly in Slack.

## Common Commands

### Build
```bash
make build
```

### Run
```bash
export SLACK_BOT_TOKEN='xoxb-...'
export SLACK_APP_TOKEN='xapp-...'
go run main.go
```

### Docker Build
```bash
docker build -t mirrorbot .
```

### Lint
```bash
make lint
```

## Architecture

- **`main.go`** - Application entry point and Slack bot logic
- **`pkg/`** - Reusable packages for OCP mirror interactions
- **`Dockerfile`** - Container image definition

## Environment Variables

| Variable | Description |
|----------|-------------|
| `SLACK_BOT_TOKEN` | Slack Bot Token (xoxb-...) |
| `SLACK_APP_TOKEN` | Slack App Token with Socket Mode (xapp-...) |

## Requirements

- Go 1.25+
- [Slack App](https://api.slack.com/apps) with Socket Mode enabled
- Docker (for containerized runs)

## Code Style

- Follow standard Go conventions
- Use `go fmt` before committing
- Run `golangci-lint` for linting
