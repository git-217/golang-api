package config

import (
	"os"
)

type Config struct {
	env      string   `env:"env" env-default:"local"`
	psqlConn pgParams `env:"psql_params"`
}

type pgParams struct {
	user     string `env:"pg_user" env-default:"postgres"`
	password string `env:"pg_password" env-default:"1111"`
	dbName   string `env:"pg_db" env-required:"true"`
	port     string `env:"pg_port"`
}

func MustLoad() *Config {
	cfg_path := os.Getenv()
}
