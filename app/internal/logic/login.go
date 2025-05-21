package logic

import (
	"social-network/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (l *Logic) Login(r models.LoginRequest) (string, error) {
	if (len(r.Username) > logic.UsernameMaxLen) {
		return "", fmt.Errorf("username too long")
	}
	if (r.Password == "") {
		return "", fmt.Errorf("password not set") 
	}
	
	user, err := l.storage.GetUserByName(r.Username)
	if err != nil {
		return "", fmt.Errorf("no such user")
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(r.Password))
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}
	
	token, err := l.session.GenerateJWT(models.JWTtoken{ID: user.ID, Username: r.Username})
	if err != nil {
		return "", fmt.Errorf("can't generate jwt token")
	}
	return token, nil
}
