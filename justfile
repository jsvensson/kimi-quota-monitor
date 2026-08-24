# Build the binary.
build:
    go build -o kimi-quota-monitor ./cmd/kimi-quota-monitor

# Run tests and vet.
test:
    go vet ./...
    go test ./...

# Build and run the service. Requires KIMI_API_KEY and MQTT_BROKER in the environment.
run: build
    ./kimi-quota-monitor

# Remove build artifacts.
clean:
    rm -f kimi-quota-monitor
