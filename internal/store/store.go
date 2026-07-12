package store

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/davralin/alertmanager-dashboard/internal/alertmanager"
	"github.com/redis/go-redis/v9"
)

const (
	lastPingKey = "alertmanager-dashboard:last_ping"
	alertsKey   = "alertmanager-dashboard:alerts"
)

type Store struct {
	client *redis.Client
}

type Alert struct {
	Fingerprint  string            `json:"fingerprint"`
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Receiver     string            `json:"receiver"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}

type State struct {
	LastPing      *time.Time `json:"lastPing"`
	LastPingAge   string     `json:"lastPingAge"`
	LastPingStale bool       `json:"lastPingStale"`
	Alerts        []Alert    `json:"alerts"`
}

func New(options *redis.Options) *Store {
	return &Store{client: redis.NewClient(options)}
}

func (s *Store) Close() error {
	return s.client.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *Store) ApplyWebhook(ctx context.Context, webhook alertmanager.Webhook, now time.Time) error {
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, lastPingKey, now.UTC().Format(time.RFC3339Nano), 0)

	for _, incoming := range webhook.Alerts {
		fingerprint := incoming.StableFingerprint()
		if incoming.IsResolved(now) || incoming.IsWatchdog() {
			pipe.HDel(ctx, alertsKey, fingerprint)
			continue
		}

		stored := Alert{
			Fingerprint:  fingerprint,
			Status:       incoming.Status,
			Labels:       incoming.Labels,
			Annotations:  incoming.Annotations,
			StartsAt:     incoming.StartsAt,
			EndsAt:       incoming.EndsAt,
			GeneratorURL: incoming.GeneratorURL,
			Receiver:     webhook.Receiver,
			UpdatedAt:    now.UTC(),
		}
		payload, err := json.Marshal(stored)
		if err != nil {
			return err
		}
		pipe.HSet(ctx, alertsKey, fingerprint, payload)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) State(ctx context.Context, now time.Time, staleAfter time.Duration) (State, error) {
	pipe := s.client.Pipeline()
	lastPingCmd := pipe.Get(ctx, lastPingKey)
	alertsCmd := pipe.HGetAll(ctx, alertsKey)
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return State{}, err
	}

	state := State{Alerts: []Alert{}}
	if lastPingRaw, err := lastPingCmd.Result(); err == nil {
		lastPing, err := time.Parse(time.RFC3339Nano, lastPingRaw)
		if err == nil {
			state.LastPing = &lastPing
			age := now.Sub(lastPing)
			if age < 0 {
				age = 0
			}
			state.LastPingAge = age.Round(time.Second).String()
			state.LastPingStale = age > staleAfter
		}
	}

	for fingerprint, raw := range alertsCmd.Val() {
		var alert Alert
		if err := json.Unmarshal([]byte(raw), &alert); err != nil {
			continue
		}
		if alert.Fingerprint == "" {
			alert.Fingerprint = fingerprint
		}
		state.Alerts = append(state.Alerts, alert)
	}

	sort.Slice(state.Alerts, func(i, j int) bool {
		left, right := state.Alerts[i], state.Alerts[j]
		if left.Labels["severity"] != right.Labels["severity"] {
			return severityRank(left.Labels["severity"]) < severityRank(right.Labels["severity"])
		}
		if !left.StartsAt.Equal(right.StartsAt) {
			return left.StartsAt.Before(right.StartsAt)
		}
		return left.Fingerprint < right.Fingerprint
	})

	return state, nil
}

func severityRank(severity string) int {
	switch severity {
	case "critical":
		return 0
	case "warning":
		return 1
	case "info":
		return 2
	default:
		return 3
	}
}
