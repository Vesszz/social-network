package handlers

import (
	"net/http"
	"social-network/internal/logic"
	"social-network/internal/middleware"
)

type Handlers struct {
	logic *logic.Logic
}

func New(l *logic.Logic) (*Handlers, error) {
	return &Handlers{logic: l}
}

func (h *Handlers) RegisterRoutes() {
	http.HandleFunc("/", h.index)
	http.HandleFunc("/register", h.register)
	http.HandleFunc("/login", h.login)
	http.HandleFunc("/post", middleware.Auth(h.post))
	http.HandleFunc("/feed", middleware.Auth(h.feed))
	http.HandleFunc("/profile/", h.profile)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}
