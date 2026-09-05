// Command kimi-quota-monitor polls the Kimi Code quota API and publishes
// the used/limit pairs to an MQTT broker. It can also serve the same
// payload over HTTP.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jsvensson/kimi-quota-monitor/internal/config"
	"github.com/jsvensson/kimi-quota-monitor/internal/httpapi"
	"github.com/jsvensson/kimi-quota-monitor/internal/kimi"
	"github.com/jsvensson/kimi-quota-monitor/internal/mqtt"
	"github.com/jsvensson/kimi-quota-monitor/internal/payload"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	level := new(slog.LevelVar)
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return fmt.Errorf("invalid LOG_LEVEL: %w", err)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	pub, err := mqtt.NewPublisher(cfg.MQTTBroker, cfg.MQTTClientID, cfg.MQTTTopic, cfg.MQTTUsername, cfg.MQTTPassword)
	if err != nil {
		return err
	}
	defer pub.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := kimi.NewClient(cfg.BaseURL, cfg.APIKey)

	srv := httpapi.NewServer(cfg.HTTPAddr)
	if len(cfg.HTTPAddr) > 0 {
		go func() {
			if err := srv.Run(ctx); err != nil {
				slog.Error("HTTP server", "error", err)
			}
		}()
	}

	slog.Info("started",
		"broker", cfg.MQTTBroker,
		"topic", cfg.MQTTTopic,
		"interval", cfg.PollInterval,
		"http_addr", cfg.HTTPAddr,
	)

	publish(ctx, client, pub, srv)

	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopped")
			return nil
		case <-ticker.C:
			publish(ctx, client, pub, srv)
		}
	}
}

// publish fetches the quota and publishes it to MQTT and the HTTP server.
// Errors are logged; the poll loop continues.
func publish(ctx context.Context, client *kimi.Client, pub *mqtt.Publisher, srv *httpapi.Server) {
	usages, err := client.Usages(ctx)
	if err != nil {
		slog.Warn("fetch quota", "error", err)
		return
	}

	data, err := payload.FromUsages(usages)
	if err != nil {
		slog.Warn("build payload", "error", err)
		return
	}
	srv.SetPayload(data)

	if err := pub.Publish(data); err != nil {
		slog.Warn("publish", "error", err)
		return
	}
	slog.Info("published", "payload", string(data))
}
