package handlers

func (h *Handler) feed(w http.ResponseWriter, r *http.Request) {
	// Извлекаем имя пользователя из контекста
	username := r.Context().Value("user").(string)

	// Используем имя пользователя для запроса постов из базы данных
	rows, err := db.Query("SELECT id, content, created_at FROM posts WHERE author_id = (SELECT id FROM users WHERE username = $1) ORDER BY created_at DESC", username)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.Content, &p.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, p)
	}

	templates.ExecuteTemplate(w, "feed.html", posts)
}
