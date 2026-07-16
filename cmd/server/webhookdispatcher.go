package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// webhookJob represents a single event to be dispatched to a webhook.
type webhookJob struct {
	SessionID string
	EventType string
	Data      []byte // JSON-encoded event data
}

// WebhookDispatcher manages the dispatch of events to configured webhooks.
// It runs a background worker that reads from a job queue and sends HTTP POST
// requests to matching webhook URLs.
type WebhookDispatcher struct {
	store  *webhookStore
	client *http.Client
	log    *slog.Logger

	jobs chan webhookJob

	mu      sync.RWMutex
	webhooks []webhookRow // cached enabled webhooks, refreshed periodically
	lastRefresh time.Time
}

const (
	webhookQueueSize    = 1024
	webhookTimeout      = 30 * time.Second
	webhookMaxRetries   = 3
	webhookRefreshEvery = 30 * time.Second
)

func newWebhookDispatcher(store *webhookStore, log *slog.Logger) *WebhookDispatcher {
	return &WebhookDispatcher{
		store:  store,
		client: &http.Client{Timeout: webhookTimeout},
		log:    log.With("component", "webhook"),
		jobs:   make(chan webhookJob, webhookQueueSize),
	}
}

// Start launches the background worker goroutine.
func (d *WebhookDispatcher) Start(ctx context.Context) {
	go d.run(ctx)
}

// Enqueue adds an event to the dispatch queue. It never blocks; if the queue
// is full, the event is dropped (best-effort).
func (d *WebhookDispatcher) Enqueue(sessionID, evtType string, data []byte) {
	select {
	case d.jobs <- webhookJob{SessionID: sessionID, EventType: evtType, Data: data}:
	default:
		d.log.Warn("webhook queue full, dropping event",
			"sessionId", sessionID, "eventType", evtType)
	}
}

func (d *WebhookDispatcher) run(ctx context.Context) {
	refreshTicker := time.NewTicker(webhookRefreshEvery)
	defer refreshTicker.Stop()

	// Initial load
	d.refreshWebhooks(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case job := <-d.jobs:
			d.dispatch(ctx, job)
		case <-refreshTicker.C:
			d.refreshWebhooks(ctx)
		}
	}
}

func (d *WebhookDispatcher) refreshWebhooks(ctx context.Context) {
	whs, err := d.store.listAllEnabled(ctx)
	if err != nil {
		d.log.Error("failed to refresh webhooks", "error", err)
		return
	}
	d.mu.Lock()
	d.webhooks = whs
	d.lastRefresh = time.Now()
	d.mu.Unlock()
	d.log.Debug("webhooks refreshed", "count", len(whs))
}

// matchingWebhooks returns webhooks that match the given session and event type.
func (d *WebhookDispatcher) matchingWebhooks(sessionID, evtType string) []webhookRow {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// If webhooks haven't been refreshed yet, do a quick check
	if d.webhooks == nil {
		return nil
	}

	var matched []webhookRow
	for _, wh := range d.webhooks {
		if wh.SessionID != sessionID {
			continue
		}
		if !wh.Enabled {
			continue
		}
		if wh.Events == "*" || matchesEvent(wh.Events, evtType) {
			matched = append(matched, wh)
		}
	}
	return matched
}

// matchesEvent checks if the event type is in the comma-separated list.
func matchesEvent(eventList, evtType string) bool {
	for _, e := range strings.Split(eventList, ",") {
		if strings.TrimSpace(e) == evtType {
			return true
		}
	}
	return false
}

func (d *WebhookDispatcher) dispatch(ctx context.Context, job webhookJob) {
	webhooks := d.matchingWebhooks(job.SessionID, job.EventType)
	if len(webhooks) == 0 {
		return
	}

	for _, wh := range webhooks {
		d.sendWithRetry(ctx, wh, job)
	}
}

func (d *WebhookDispatcher) sendWithRetry(ctx context.Context, wh webhookRow, job webhookJob) {
	// Build the payload
	payload := map[string]any{
		"event":     job.EventType,
		"sessionId": job.SessionID,
		"timestamp": time.Now().UnixMilli(),
	}

	// Unmarshal the original data to embed it
	var rawData any
	if err := json.Unmarshal(job.Data, &rawData); err == nil {
		payload["data"] = rawData
	}

	body, err := json.Marshal(payload)
	if err != nil {
		d.log.Error("failed to marshal webhook payload", "error", err)
		return
	}

	for attempt := 1; attempt <= webhookMaxRetries; attempt++ {
		if attempt > 1 {
			// Exponential backoff: 1s, 3s, 5s
			backoff := time.Duration(2*attempt-1) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return
			}
		}

		err := d.Send(ctx, wh, body)
		if err == nil {
			return
		}

		d.log.Warn("webhook dispatch failed",
			"webhookId", wh.ID,
			"url", wh.URL,
			"eventType", job.EventType,
			"attempt", attempt,
			"error", err)
	}

	d.log.Error("webhook dispatch failed after retries",
		"webhookId", wh.ID,
		"url", wh.URL,
		"eventType", job.EventType)
}

// Send delivers the payload to a single webhook URL.
// Exported so the test endpoint can call it directly.
func (d *WebhookDispatcher) Send(ctx context.Context, wh webhookRow, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "NB_API-Webhook/1.0")
	req.Header.Set("X-Webhook-Event", wh.Events)

	// Sign the body if a secret is configured
	if wh.Secret != "" {
		mac := hmac.New(sha256.New, []byte(wh.Secret))
		mac.Write(body)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Webhook-Signature", signature)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

// Sink returns a function that can be set as Broker.WebhookSink.
// It unmarshals the event JSON to extract sessionId and type, then enqueues.
func (d *WebhookDispatcher) Sink() func(data []byte) {
	return func(data []byte) {
		// Quick parse to extract sessionId and type
		var env struct {
			Type      string `json:"type"`
			SessionID string `json:"sessionId"`
		}
		if err := json.Unmarshal(data, &env); err != nil {
			return
		}
		if env.SessionID == "" {
			return
		}
		d.Enqueue(env.SessionID, env.Type, data)
	}
}
