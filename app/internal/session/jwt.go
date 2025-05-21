package session

import (
	"github.com/golang-jwt/jwt/v5"
	"social-network/internal/models"
	"time"
)

func (s *Session) GenerateJWT(jwttoken *models.JWTtoken) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": jwttoken.ID,
		"username": jwttoken.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(s.JWTKey)
}

func (s *Session) ParseJWT(tokenString string) (*models.JWTtoken, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return s.JWTKey, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	id, ok := claims["id"].(int)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	exp, ok := claims["exp"].(time.Time)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	return &models.JWTtoken{ID: id, Username: username, Exp: exp}, nil
}
