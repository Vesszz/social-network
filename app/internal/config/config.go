package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type DBConfig struct {
	Host string `env:"DB_HOST,required"`
	User string `env:"DB_USER,required"`
	Password string `env:"DB_PASSWORD,required"`
	Name string `env:"DB_NAME,required"`
	Conn string
}

type Config struct {
	DB DBConfig
}

func initDBConfig() (DBConfig, error) {
	var dbCfg DBConfig
	if err := env.Parse(&dbCfg); err != nil {
		return DBConfig{}, fmt.Errorf("parse from .env to struct: %w", err)
	}

	dbCfg.Conn = fmt.Sprintf(
		"host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
		dbCfg.Host,
		dbCfg.User,
		dbCfg.Password,
		dbCfg.Name,
	)

	return dbCfg, nil
}

func InitConfig() (*Config, error) {
	dbCfg, err := initDBConfig()
	if err != nil {
		return nil, fmt.Errorf("init DB config: %w", err)
	}
	return &Config{
		DB:  dbCfg
	}, nil
}
