package middleware

import (
	"context"
	"net/http"
	"social-network/internal/session"
)

func Auth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r * http.Request) {
        cookie, err := r.Cookie("jwt")
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        tokenString := cookie.Value
        username, err := session.ParseJWT(tokenString)
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        ctx := context.WithValue(r.Context(), "user", username)
        next(w, r.WithContext(ctx))
    }
}
