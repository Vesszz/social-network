package handlers

import (
	"net/http"
	"social-network/internal/storage"
	"social-network/internal/middleware"
)

type Handlers struct {
	storage storage.Storage
}

func New() (*Handlers, error) {
	return &Handlers{storage: storage}
}

func (h *Handlers) RegisterRoutes() {
	http.HandleFunc("/", h.indexHandler)
	http.HandleFunc("/register", h.registerHandler)
	http.HandleFunc("/login", h.loginHandler)
	http.HandleFunc("/post", middleware.Auth(h.postHandler))
	http.HandleFunc("/feed", middleware.Auth(h.feedHandler))
	http.HandleFunc("/profile/", h.profileHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}
