package handlers

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		http.Redirect(w, r, "/profile/"+username, http.StatusFound)
		return
	}

	// Запрос для получения последних 5 постов
	rows, err := db.Query(`
        SELECT p.id, p.content, p.created_at, u.username
        FROM posts p
        JOIN users u ON p.author_id = u.id
        ORDER BY p.created_at DESC
        LIMIT 5
    `)
	if err != nil {
		log.Printf("Error fetching latest posts: %v", err)
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []struct {
		ID        int
		Content   string
		CreatedAt time.Time
		Author    string
	}

	for rows.Next() {
		var p struct {
			ID        int
			Content   string
			CreatedAt time.Time
			Author    string
		}
		err := rows.Scan(&p.ID, &p.Content, &p.CreatedAt, &p.Author)
		if err != nil {
			log.Printf("Error scanning post: %v", err)
			http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, p)
	}

	// Передаем посты в шаблон
	templates.ExecuteTemplate(w, "index.html", posts)
}
