package main

import (
	"context"
	"log"
	"net/http"

	"profile-svc/internal/config"
	"profile-svc/internal/handlers"
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

	profileStore := store.NewProfileStore(db)

	profileService := services.NewProfileService(profileStore)


	server := handlers.NewServer(profileService)

	log.Printf("listening on %s", cfg.ServerAddr)
	if err := http.ListenAndServe(cfg.ServerAddr, server.Router()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
