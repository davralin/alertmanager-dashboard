package store

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/davralin/alertmanager-dashboard/internal/alertmanager"
	"github.com/redis/go-redis/v9"
)

func TestApplyWebhookStoresAndResolvesAlerts(t *testing.T) {
	server := miniredis.RunT(t)
	st := New(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = st.Close() })

	ctx := context.Background()
	now := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC)
	alert := alertmanager.Alert{
		Status:      "firing",
		Labels:      map[string]string{"alertname": "NodeDown", "severity": "critical", "instance": "node-1"},
		Annotations: map[string]string{"summary": "node is down"},
		StartsAt:    now.Add(-time.Hour),
		Fingerprint: "abc123",
	}

	if err := st.ApplyWebhook(ctx, alertmanager.Webhook{Receiver: "dashboard", Alerts: []alertmanager.Alert{alert}}, now); err != nil {
		t.Fatalf("apply firing webhook: %v", err)
	}

	state, err := st.State(ctx, now.Add(time.Minute), 90*time.Minute)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if state.LastPing == nil || !state.LastPing.Equal(now) {
		t.Fatalf("last ping = %v, want %v", state.LastPing, now)
	}
	if len(state.Alerts) != 1 {
		t.Fatalf("alerts len = %d, want 1", len(state.Alerts))
	}
	if state.Alerts[0].Fingerprint != "abc123" {
		t.Fatalf("fingerprint = %q, want abc123", state.Alerts[0].Fingerprint)
	}

	alert.Status = "resolved"
	alert.EndsAt = now.Add(time.Minute)
	if err := st.ApplyWebhook(ctx, alertmanager.Webhook{Receiver: "dashboard", Alerts: []alertmanager.Alert{alert}}, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("apply resolved webhook: %v", err)
	}

	state, err = st.State(ctx, now.Add(3*time.Minute), 90*time.Minute)
	if err != nil {
		t.Fatalf("load resolved state: %v", err)
	}
	if len(state.Alerts) != 0 {
		t.Fatalf("alerts len = %d, want 0", len(state.Alerts))
	}
}

func TestApplyWebhookUsesWatchdogOnlyForLastPing(t *testing.T) {
	server := miniredis.RunT(t)
	st := New(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = st.Close() })

	ctx := context.Background()
	now := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC)
	watchdog := alertmanager.Alert{
		Status:      "firing",
		Labels:      map[string]string{"alertname": "Watchdog", "severity": "none"},
		Annotations: map[string]string{"summary": "heartbeat"},
		StartsAt:    now.Add(-24 * time.Hour),
		Fingerprint: "watchdog",
	}

	if err := st.ApplyWebhook(ctx, alertmanager.Webhook{Receiver: "dashboard", Alerts: []alertmanager.Alert{watchdog}}, now); err != nil {
		t.Fatalf("apply watchdog webhook: %v", err)
	}

	state, err := st.State(ctx, now.Add(2*time.Hour), 90*time.Minute)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if state.LastPing == nil || !state.LastPing.Equal(now) {
		t.Fatalf("last ping = %v, want %v", state.LastPing, now)
	}
	if !state.LastPingStale {
		t.Fatal("last ping should be stale after two hours")
	}
	if len(state.Alerts) != 0 {
		t.Fatalf("alerts len = %d, want 0", len(state.Alerts))
	}
}
