package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
	_ "github.com/lib/pq"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type server struct {
	broker       *Broker
	sessions     *SessionManager
	log          *slog.Logger
	staticDir    string
	redis        *redis.Client
	authUsername string
	authPassword string
	natIP        string
	swaggerURL   string
	webhookStore      *webhookStore
	webhookDispatcher *WebhookDispatcher
	callHistory       *callHistoryStore
}

func openDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// Configure pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return db, nil
}

func initRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	return rdb
}

func newServer(ctx context.Context, dbPath, staticDir string, maxCalls int, swaggerURL string, log *slog.Logger) (*server, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	container := sqlstore.NewWithDB(db, "postgres", waLog.Noop)
	if err := container.Upgrade(ctx); err != nil {
		return nil, err
	}
	store, err := newSessionStore(ctx, db)
	if err != nil {
		return nil, err
	}

	waLogger := waLog.Noop
	if log.Enabled(ctx, slog.LevelDebug) {
		waLogger = waLog.Stdout("WA", "INFO", true)
	}

	rdb := initRedis()

	broker := NewBroker()
	mgr := newSessionManager(ctx, container, broker, store, waLogger, log, maxCalls)
	broker.SnapshotFn = mgr.snapshotEvents

	whStore, err := newWebhookStore(ctx, db)
	if err != nil {
		return nil, err
	}
	whDispatcher := newWebhookDispatcher(whStore, log)
	broker.WebhookSink = whDispatcher.Sink()
	whDispatcher.Start(ctx)

	callHistory, err := newCallHistoryStore(ctx, db)
	if err != nil {
		return nil, err
	}
	broker.HistoryStore = callHistory

	return &server{
		broker:            broker,
		sessions:          mgr,
		log:               log,
		staticDir:         staticDir,
		redis:             rdb,
		authUsername:      os.Getenv("AUTH_USERNAME"),
		authPassword:      os.Getenv("AUTH_PASSWORD"),
		natIP:             os.Getenv("EXTERNAL_IP"),
		swaggerURL:        swaggerURL,
		webhookStore:      whStore,
		webhookDispatcher: whDispatcher,
		callHistory:       callHistory,
	}, nil
}
