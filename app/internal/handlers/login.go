package handlers

import (
	"database/sql"
	"social-network/internal/session"
)

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var storedHash string
		err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&storedHash)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		token, err := session.GenerateJWT(username)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
		})
		http.Redirect(w, r, "/feed", http.StatusFound)
	} else {
		templates.ExecuteTemplate(w, "login.html", nil)
	}
}
