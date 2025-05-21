package session

import (
	"social-network/internal/config"
)

type Session struct {
	JWTKey string
}

func New(jwtCfg config.JWTConfig) (*Session, error) {
	return &Session{JWTKey: jwtCfg.Key}, nil
}
