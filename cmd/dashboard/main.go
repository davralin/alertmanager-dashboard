package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davralin/alertmanager-dashboard/internal/config"
	"github.com/davralin/alertmanager-dashboard/internal/healthcheck"
	"github.com/davralin/alertmanager-dashboard/internal/store"
)

var page = template.Must(template.New("dashboard").Funcs(template.FuncMap{
	"label": func(labels map[string]string, key string) string { return labels[key] },
	"formatTime": func(ts time.Time) string {
		if ts.IsZero() {
			return ""
		}
		return ts.UTC().Format("2006-01-02 15:04:05 UTC")
	},
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="refresh" content="30">
  <title>Alertmanager Dashboard</title>
  <style>
    :root { color-scheme: dark; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #0f172a; color: #e2e8f0; }
    body { margin: 0; padding: 2rem; }
    main { max-width: 1200px; margin: 0 auto; }
    header { display: flex; flex-wrap: wrap; gap: 1rem; align-items: end; justify-content: space-between; margin-bottom: 2rem; }
    h1 { margin: 0; font-size: clamp(2rem, 5vw, 4rem); line-height: 1; letter-spacing: -0.05em; }
    .summary { display: flex; flex-wrap: wrap; gap: 0.75rem; }
    .pill { border: 1px solid #334155; border-radius: 999px; padding: 0.5rem 0.75rem; background: #111827; }
    .stale { border-color: #f97316; color: #fed7aa; }
    .fresh { border-color: #22c55e; color: #bbf7d0; }
    .notice { border: 1px solid #f97316; border-radius: 1rem; padding: 1rem; color: #fed7aa; background: #431407; margin-bottom: 1rem; }
    .empty { border: 1px dashed #475569; border-radius: 1rem; padding: 3rem; text-align: center; color: #94a3b8; }
    .alerts { display: grid; gap: 1rem; }
    article { border: 1px solid #334155; border-radius: 1rem; background: linear-gradient(135deg, #111827, #0f172a); padding: 1rem; box-shadow: 0 16px 40px rgb(0 0 0 / 0.25); }
    article.critical { border-color: #ef4444; }
    article.warning { border-color: #f59e0b; }
    article.info { border-color: #38bdf8; }
    h2 { margin: 0 0 0.75rem; font-size: 1.2rem; }
    dl { display: grid; grid-template-columns: max-content 1fr; gap: 0.35rem 0.75rem; margin: 0.75rem 0; }
    dt { color: #94a3b8; }
    dd { margin: 0; word-break: break-word; }
    a { color: #93c5fd; }
    code { color: #c4b5fd; }
    @media (max-width: 700px) { body { padding: 1rem; } dl { grid-template-columns: 1fr; } dt { margin-top: 0.5rem; } }
  </style>
</head>
<body>
<main>
  <header>
    <div>
      <h1>Alerts</h1>
      <p>Active Alertmanager alerts.</p>
    </div>
    <div class="summary">
      <span class="pill">Active: {{ len .Alerts }}</span>
      {{ if .LastUpdate }}<span class="pill {{ if .LastUpdateStale }}stale{{ else }}fresh{{ end }}">Last update: {{ .LastUpdateAge }} ago</span>{{ else }}<span class="pill stale">No updates received</span>{{ end }}
      <a class="pill" href="/api/state">JSON</a>
    </div>
  </header>
  {{ if .LastUpdate }}{{ if .LastUpdateStale }}<section class="notice">No Alertmanager updates have been received recently. Alert data may be stale.</section>{{ end }}{{ else }}<section class="notice">No Alertmanager updates have been received yet.</section>{{ end }}
  {{ if .Alerts }}
    <section class="alerts">
    {{ range .Alerts }}
      <article class="{{ label .Labels "severity" }}">
        <h2>{{ label .Labels "alertname" }}</h2>
        <dl>
          {{ with label .Labels "severity" }}<dt>Severity</dt><dd>{{ . }}</dd>{{ end }}
          {{ with label .Labels "namespace" }}<dt>Namespace</dt><dd>{{ . }}</dd>{{ end }}
          {{ with label .Labels "instance" }}<dt>Instance</dt><dd>{{ . }}</dd>{{ end }}
          {{ with label .Annotations "summary" }}<dt>Summary</dt><dd>{{ . }}</dd>{{ end }}
          {{ with label .Annotations "description" }}<dt>Description</dt><dd>{{ . }}</dd>{{ end }}
          <dt>Started</dt><dd>{{ formatTime .StartsAt }}</dd>
          <dt>Updated</dt><dd>{{ formatTime .UpdatedAt }}</dd>
          <dt>Receiver</dt><dd>{{ .Receiver }}</dd>
          <dt>Fingerprint</dt><dd><code>{{ .Fingerprint }}</code></dd>
          {{ if .GeneratorURL }}<dt>Source</dt><dd><a href="{{ .GeneratorURL }}">generator URL</a></dd>{{ end }}
        </dl>
      </article>
    {{ end }}
    </section>
  {{ else }}
    {{ if .LastUpdate }}{{ if .LastUpdateStale }}<section class="empty">No active alerts shown, but alert updates may be stale.</section>{{ else }}<section class="empty">No active alerts.</section>{{ end }}{{ else }}<section class="empty">No alert data received yet.</section>{{ end }}
  {{ end }}
</main>
</body>
</html>`))

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
		Handler:           routes(st, cfg.StaleAfter),
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

	slog.Info("starting dashboard", "addr", cfg.ListenAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func routes(st *store.Store, staleAfter time.Duration) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := st.Ping(ctx); err != nil {
			http.Error(w, "state store unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/state", stateHandler(st, staleAfter))
	mux.HandleFunc("GET /", dashboardHandler(st, staleAfter))
	return mux
}

func dashboardHandler(st *store.Store, staleAfter time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		state, err := st.State(ctx, time.Now(), staleAfter)
		if err != nil {
			slog.Error("failed to load state", "error", err)
			http.Error(w, "failed to load state", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := page.Execute(w, state); err != nil {
			slog.Error("failed to render dashboard", "error", err)
		}
	}
}

func stateHandler(st *store.Store, staleAfter time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		state, err := st.State(ctx, time.Now(), staleAfter)
		if err != nil {
			slog.Error("failed to load state", "error", err)
			http.Error(w, "failed to load state", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(state)
	}
}
