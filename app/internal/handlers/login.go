package handlers

import (
	"social-network/internal/session"
	"social-network/internal/models"
	"social-network/internal/logic"
)

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		jwttoken, err := h.logic.Login(models.LoginRequest{Username: username, Password: password})
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusBadRequest)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    jwttoken,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
		})
		http.Redirect(w, r, "/feed", http.StatusFound)
	} else if r.Method == "GET" {
		templates.ExecuteTemplate(w, "login.html", nil)
		return
	} else {
		http.Error(w, "Unsupported request method", http.StatusMethodNotAllowed)
		return
	}
}
