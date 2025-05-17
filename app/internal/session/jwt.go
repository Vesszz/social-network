package session

import (
	"github.com/golang-jwt/jwt/v5"
	"log"
	"os"
	"time"
)

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
func ParseJWT(tokenString string) (string, error) {
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
