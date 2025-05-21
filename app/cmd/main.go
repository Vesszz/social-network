package main

import (
	"log"
	"net/http"
	"social-network/internal/config"
	"social-network/internal/handlers"
	"social-network/internal/session"
	"social-network/internal/storage"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("init config: %w", err)	
	}
	storage, err := storage.New(cfg.DB.Conn)
	if err != nil {
		log.Fatalf("init storage: %w", err)
	}
	defer storage.Close()
	session, err := session.New(cfg.JWT)
	if err != nil {
		log.Fatalf("init session: %w", err)
	}
	logic, err := logic.New(storage, session)
	if err != nil {
		log.Fatalf("init logic: %w", err)
	}
	handlers, err := handlers.New(logic)
	if err != nil {
		log.Fatalf("init handlers: %w", err)
	}
	handlers.RegisterRoutes()
}
