package logic

import (
	"social-network/internal/models"
)

const (
	UsernameMaxLen = 50
)

type Storage interface {
	GetUsers() ([]models.User, error)
	CreateUser(*models.User) (int, error)
	GetUserByName(string) (models.User, error)
	DeleteUser(u *models.User) error
}

type Session interface {
	GenerateJWT(string) (string, error)
	ParseJWT(string) (string, error)
}

type Logic struct {
	storage Storage
	session Session
}

func New(st Storage, se Session) (*Logic, error) {
	return &Logic{storage: st, session: se}, nil
} 
