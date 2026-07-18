package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"profile-svc/internal/config"
	"profile-svc/internal/handlers"
	"profile-svc/internal/messaging"
	"profile-svc/internal/services"
	"profile-svc/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	db, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("database is not reachable: %v", err)
	}
	log.Println("connected to database")

	broker, err := messaging.InitBroker(cfg)
	if err != nil {
		log.Fatalf("failed to init broker: %v", err)
	}

	profileStore := store.NewProfileStore(db)
	profileService := services.NewProfileService(profileStore, broker)
	server := handlers.NewServer(profileService, broker)

	go runHttpServer(cfg.ServerAddr, server.Router())

	slog.Info("starting broker loop")
	broker.Run()
	slog.Info("broker loop exited")
}

func runHttpServer(s string, handler http.Handler) {
	if err := http.ListenAndServe(s, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	} else {
		log.Println("http server started")
	}
}
