package main

import (
	"log"
	"net/http"
	"social-network/internal/config"
	"social-network/internal/handlers"
	"social-network/internal/middleware"
	"social-network/internal/models"
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
	handlers, err := handlers.New(storage)
	if err != nil {
		log.Fatalf("init handlers: %w", err)
	}
}
