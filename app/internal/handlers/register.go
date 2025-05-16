package handlers

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		hashedPass, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		_, err := db.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)", username, string(hashedPass))
		if err != nil {
			if strings.Contains(err.Error(), "unique") {
				http.Error(w, "Username already exists", http.StatusBadRequest)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusFound)
	} else {
		templates.ExecuteTemplate(w, "register.html", nil)
	}
}
