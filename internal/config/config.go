package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string   `yaml:"env" env-default:"local"`
	PsqlConn pgParams `yaml:"psql_params"`
}

type pgParams struct {
	User     string `yaml:"pg_user" env-default:"postgres"`
	Password string `yaml:"pg_password" env-default:"1111"`
	DbName   string `yaml:"pg_db" env-required:"true"`
	Port     string `yaml:"pg_port" env-default:"5432"`
}

func MustLoad() *Config {
	configPath := os.Getenv("PSQLCRUD_CFG")
	if configPath == "" {
		log.Fatal("PSQLCRUD_CFG is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Can't read config file: %s", configPath)
	}

	return &cfg
}
