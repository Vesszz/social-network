package middleware

import (
	"net/http"
	"social-network/internal/session"
)

func Auth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r * http.Request) {
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
