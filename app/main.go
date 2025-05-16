package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Username string
}

type Post struct {
	ID        int
	Content   string
	AuthorID  int
	CreatedAt time.Time
}

var db *sql.DB
var templates = template.Must(template.ParseGlob("templates/*.html"))

const sessionDuration = 24 * time.Hour

func initDB() {
	var err error
	connStr := fmt.Sprintf(
		"host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to the database")
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/post", authMiddleware(postHandler))
	http.HandleFunc("/feed", authMiddleware(feedHandler))
	http.HandleFunc("/profile/", profileHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server started on :8080")
	http.ListenAndServe("0.0.0.0:8080", nil)
}

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

func loginHandler(w http.ResponseWriter, r *http.Request) {
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
		token, err := generateJWT(username)
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

func feedHandler(w http.ResponseWriter, r *http.Request) {
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
func profileHandler(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/profile/")

	var user User
	err := db.QueryRow("SELECT id, username FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT id, content, created_at FROM posts WHERE author_id = $1 ORDER BY created_at DESC", user.ID)
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
	templates.ExecuteTemplate(w, "profile.html", struct {
		User  User
		Posts []Post
	}{
		User:  user,
		Posts: posts,
	})
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		tokenString := cookie.Value
		username, err := parseJWT(tokenString)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), "user", username)
		next(w, r.WithContext(ctx))
	}
}

func getJWTKey() []byte {
	jwtKey := os.Getenv("JWT_SECRET_KEY")
	if jwtKey == "" {
		log.Fatal("JWT_SECRET_KEY is not set in environment variables")
	}
	return []byte(jwtKey)
}

// Создание JWT-токена
func generateJWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // Токен действителен 24 часа
	})

	return token.SignedString(getJWTKey())
}

// Проверка JWT-токена
func parseJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrInvalidKey
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", jwt.ErrInvalidKey
	}

	return username, nil
}
