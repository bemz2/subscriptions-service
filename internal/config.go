package internal

type AppConfig struct {
	PostgresConfig PostgresConfig
}

type PostgresConfig struct {
	DBAdapter  string `env:"DB_ADAPTER"`
	DBName     string `env:"POSTGRES_DB"`
	DBHost     string `env:"POSTGRES_HOST"`
	DBPort     string `env:"POSTGRES_PORT"`
	DBUser     string `env:"POSTGRES_USER"`
	DBPassword string `env:"POSTGRES_PASSWORD"`
	DBSSLMode  string `env:"POSTGRES_SSLMODE"`
}
