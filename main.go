package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string
	Password string
	Posts    []Post
}

type Post struct {
	Content   string
	Author    string
	Timestamp time.Time
}

type Session struct {
	Username string
	Expiry   time.Time
}

const sessionDuration = 24 * time.Hour

var (
	users    = make(map[string]User)
	sessions = make(map[string]Session)
	posts    []Post
	mutex    sync.RWMutex
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/post", authMiddleware(postHandler))
	http.HandleFunc("/feed", authMiddleware(feedHandler))
	http.HandleFunc("/profile/", profileHandler) // Открытый доступ к профилям
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		mutex.Lock()
		defer mutex.Unlock()

		if _, exists := users[username]; exists {
			http.Error(w, "Username exists", http.StatusBadRequest)
			return
		}

		hashedPass, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		// Создаем нового пользователя
		newUser := User{
			Username: username,
			Password: string(hashedPass),
			Posts:    []Post{},
		}
		// Сохраняем в мапу
		users[username] = newUser

		http.Redirect(w, r, "/login", http.StatusFound)
	} else {
		templates.ExecuteTemplate(w, "register.html", nil)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		mutex.RLock()
		user, exists := users[username]
		mutex.RUnlock()

		if !exists || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		sessionToken := generateSessionToken()
		expiresAt := time.Now().Add(sessionDuration)

		mutex.Lock()
		sessions[sessionToken] = Session{Username: username, Expiry: expiresAt}
		mutex.Unlock()

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  expiresAt,
			HttpOnly: true,
			Path:     "/",
		})

		http.Redirect(w, r, "/feed", http.StatusFound)
	} else {
		templates.ExecuteTemplate(w, "login.html", nil)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		content := r.FormValue("content")
		author := r.Context().Value("user").(string)

		newPost := Post{
			Content:   content,
			Author:    author,
			Timestamp: time.Now(),
		}

		mutex.Lock()
		// Получаем копию пользователя
		user := users[author]
		// Добавляем пост в копию
		user.Posts = append(user.Posts, newPost)
		// Сохраняем измененного пользователя обратно в мапу
		users[author] = user
		// Добавляем в глобальные посты
		posts = append(posts, newPost)
		mutex.Unlock()

		http.Redirect(w, r, "/feed", http.StatusFound)
	}
}
func feedHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("user").(string)
	var userPosts []Post

	mutex.RLock()
	for _, post := range users[username].Posts {
		userPosts = append(userPosts, post)
	}
	mutex.RUnlock()

	templates.ExecuteTemplate(w, "feed.html", userPosts)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	profileUser := strings.TrimPrefix(r.URL.Path, "/profile/")

	mutex.RLock()
	user, exists := users[profileUser]
	mutex.RUnlock()

	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	templates.ExecuteTemplate(w, "profile.html", struct {
		User  User
		Posts []Post
	}{
		User:  user,
		Posts: user.Posts,
	})
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		mutex.RLock()
		session, exists := sessions[c.Value]
		mutex.RUnlock()

		if !exists || session.Expiry.Before(time.Now()) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), "user", session.Username)
		next(w, r.WithContext(ctx))
	}
}

func generateSessionToken() string {
	b := make([]byte, 32)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)
}
