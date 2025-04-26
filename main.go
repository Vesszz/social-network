package main

import (
    "context"
    "database/sql"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "strings"
    "time"
    "os"

    "golang.org/x/crypto/bcrypt"

    _ "github.com/lib/pq"
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
    http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        username := r.FormValue("username")
        http.Redirect(w, r, "/profile/"+username, http.StatusFound)
        return
    }
    templates.ExecuteTemplate(w, "index.html", nil)
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

        http.Redirect(w, r, "/feed", http.StatusFound)
    } else {
        templates.ExecuteTemplate(w, "login.html", nil)
    }
}

func postHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        content := r.FormValue("content")
        author := r.Context().Value("user").(User)

        _, err := db.Exec("INSERT INTO posts (content, author_id) VALUES ($1, $2)", content, author.ID)
        if err != nil {
            http.Error(w, "Failed to create post", http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, "/feed", http.StatusFound)
    }
}

func feedHandler(w http.ResponseWriter, r *http.Request) {
    user := r.Context().Value("user").(User)

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
        session, _ := r.Cookie("session_token")
        if session == nil {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        var user User
        err := db.QueryRow("SELECT id, username FROM users WHERE username = $1", session.Value).Scan(&user.ID, &user.Username)
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        ctx := context.WithValue(r.Context(), "user", user)
        next(w, r.WithContext(ctx))
    }
}
