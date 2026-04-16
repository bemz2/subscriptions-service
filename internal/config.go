package internal

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	PublicServerConfig PublicServerConfig
	PostgresConfig     PostgresConfig
}

type PublicServerConfig struct {
	Port string `env:"SERVER_PORT" envDefault:"1323"`
}

type PostgresConfig struct {
	DBAdapter      string        `env:"DB_ADAPTER"`
	DBName         string        `env:"POSTGRES_DB"`
	DBHost         string        `env:"POSTGRES_HOST"`
	DBPort         string        `env:"POSTGRES_PORT"`
	DBUser         string        `env:"POSTGRES_USER"`
	DBPassword     string        `env:"POSTGRES_PASSWORD"`
	DBSSLMode      string        `env:"POSTGRES_SSLMODE"`
	ConnectTimeout time.Duration `env:"POSTGRES_CONNECT_TIMEOUT"`
}

func NewConfig[T any](files ...string) (T, error) {
	_ = godotenv.Overload(files...)
	cfg := *(new(T))
	return cfg, env.Parse(&cfg)
}
