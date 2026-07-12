package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davralin/alertmanager-dashboard/internal/alertmanager"
	"github.com/davralin/alertmanager-dashboard/internal/config"
	"github.com/davralin/alertmanager-dashboard/internal/healthcheck"
	"github.com/davralin/alertmanager-dashboard/internal/store"
)

const maxWebhookBytes = 1 << 20

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := healthcheck.Run(config.HealthcheckURL()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	cfg := config.Load()
	st := store.New(&cfg.Valkey)
	defer st.Close()

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           routes(st),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown failed", "error", err)
		}
	}()

	slog.Info("starting receiver", "addr", cfg.ListenAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func routes(st *store.Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := st.Ping(ctx); err != nil {
			http.Error(w, "valkey unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /webhook", webhookHandler(st))
	return mux
}

func webhookHandler(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBytes)
		defer r.Body.Close()

		var webhook alertmanager.Webhook
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&webhook); err != nil {
			http.Error(w, "invalid alertmanager webhook", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := st.ApplyWebhook(ctx, webhook, time.Now()); err != nil {
			slog.Error("failed to apply webhook", "error", err)
			http.Error(w, "failed to store webhook", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"alerts": len(webhook.Alerts),
		})
	}
}
