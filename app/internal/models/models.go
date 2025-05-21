package models

import (
	"time"
)

type User struct {
	ID       int
	Username string
	PasswordHash string
	CreatedAt time.Time
}

type Post struct {
	ID        int
	Content   string
	AuthorID  int
	CreatedAt time.Time
}

type JWTtoken struct {
	ID int
	Username string
	Exp time.Time
}
