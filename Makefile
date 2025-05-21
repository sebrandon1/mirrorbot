IMAGE_NAME ?= mirrorbot:latest

.PHONY: build

build:
	go build -o mirrorbot main.go

docker-build: build
	docker build -t $(IMAGE_NAME) .

run: docker-build
	@if [ -z "$(SLACK_BOT_TOKEN)" ] || [ -z "$(SLACK_APP_TOKEN)" ]; then \
	  echo "[WARNING] SLACK_BOT_TOKEN and/or SLACK_APP_TOKEN are not set. Please supply these environment variables."; \
	  exit 1; \
	fi
	docker run --rm \
	  -e SLACK_BOT_TOKEN=$(SLACK_BOT_TOKEN) \
	  -e SLACK_APP_TOKEN=$(SLACK_APP_TOKEN) \
	  $(IMAGE_NAME)
