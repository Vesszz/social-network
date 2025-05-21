package handlers

import (
	"golang.org/x/crypto/bcrypt"
	"social-network/internal/storage"
	"social-network/internal/models"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := h.logic.Register(models.RegisterRequest{Username: username, Password: password})
		if err != nil {
			if strings.Contains(err.Error(), "username") {
				http.Errorf("Username already taken", http.StatusBadRequest)
			}
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	} else {
		templates.ExecuteTemplate(w, "register.html", nil)
	}
}
