package handlers

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Извлекаем имя пользователя из контекста
		username := r.Context().Value("user").(string)

		// Получаем содержимое поста из формы
		content := r.FormValue("content")

		// Находим ID пользователя по имени
		var authorID int
		err := db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&authorID)
		if err != nil {
			http.Error(w, "Failed to find user", http.StatusInternalServerError)
			return
		}

		// Создаем новый пост в базе данных
		_, err = db.Exec("INSERT INTO posts (content, author_id) VALUES ($1, $2)", content, authorID)
		if err != nil {
			http.Error(w, "Failed to create post", http.StatusInternalServerError)
			return
		}

		// Перенаправляем пользователя обратно в ленту
		http.Redirect(w, r, "/feed", http.StatusFound)
	}
}
