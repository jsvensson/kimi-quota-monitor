# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/kimi-quota-monitor ./cmd/kimi-quota-monitor

FROM gcr.io/distroless/static-debian12:nonroot

# Example configuration. Override at runtime with -e or --env-file.
# Required:
#   KIMI_API_KEY=sk-kimi-xxx
#   MQTT_BROKER=tcp://192.168.1.10:1883
ENV KIMI_BASE_URL="https://api.kimi.com/coding/v1" \
    POLL_INTERVAL="5m" \
    MQTT_CLIENT_ID="kimi-quota-monitor" \
    MQTT_TOPIC="quota/llm" \
    LOG_LEVEL="info"

COPY --from=build /bin/kimi-quota-monitor /kimi-quota-monitor
ENTRYPOINT ["/kimi-quota-monitor"]
