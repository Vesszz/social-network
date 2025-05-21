package logic

import (
	"social-network/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (l *Logic) Register(r models.RegisterRequest) error {
	if (len(r.Username) > logic.UsernameMaxLen) {
		return fmt.Errorf("username too long")
	}
	if (r.Password == "") {
		return fmt.Errorf("password not set")
	}
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	_, err := l.storage.CreateUser(&models.User{Username: r.Username, PasswordHash: passwordHashx})
	if err != nil {
		return fmt.Errorf("username already taken")
	}
	return nil
}
