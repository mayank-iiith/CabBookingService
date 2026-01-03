package db

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_USER" envDefault:"local_user"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"local_password"`
	DBName   string `env:"POSTGRES_DB" envDefault:"local_db"`
}
