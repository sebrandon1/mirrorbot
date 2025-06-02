# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o mirrorbot main.go

FROM --platform=$TARGETPLATFORM alpine:3.22

# Install Skopeo and ca-certificates
RUN apk add --no-cache skopeo ca-certificates

WORKDIR /app
COPY --from=builder /app/mirrorbot /app/mirrorbot
COPY --from=builder /app/pkg /app/pkg
COPY --from=builder /app/README.md /app/README.md
COPY --from=builder /app/LICENSE /app/LICENSE

ENV SLACK_BOT_TOKEN=""
ENV SLACK_APP_TOKEN=""

ENTRYPOINT ["/app/mirrorbot"]
