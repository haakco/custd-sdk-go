package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	custd "github.com/haakco/custd-sdk-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	token, err := devToken(ctx)
	if err != nil {
		fail(err)
	}

	client := custd.NewClient(&custd.ClientConfig{
		BaseURL:       envOrDefault("CUSTD_DEV_BASE_URL", "http://localhost:8087"),
		APIKey:        token,
		BatchSize:     1,
		FlushInterval: time.Hour,
	})
	// nolint:errcheck // smoke-test client teardown; a close error does not change the smoke result already reported
	defer func() { _ = client.Close(context.Background()) }()

	payload := json.RawMessage(`{"source":"sdk-go-smoke"}`)
	err = client.Track(ctx, &custd.EventEnvelope{
		EventTypeSlug: "page-view",
		SchemaVersion: "1.0.0",
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Context: custd.EventContext{
			Page:   &custd.PageContext{URL: "https://example.com"},
			Device: &custd.DeviceContext{Type: "desktop"},
		},
		Payload: payload,
	})
	if err != nil {
		fail(fmt.Errorf("custd sdk go smoke failed: %w", err))
	}

	fmt.Println("custd sdk go smoke OK")
}

func devToken(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "bash", "../../scripts/dev-hydra-token.sh")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("get hydra token: %w: %s", err, strings.TrimSpace(string(output)))
	}
	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("get hydra token: empty token")
	}
	return token, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
