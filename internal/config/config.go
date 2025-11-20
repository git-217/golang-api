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
	Host     string `yaml:"pg_host" env-default:"localhost"`
	Port     int32  `yaml:"pg_port" env-default:"5432"`
	DbName   string `yaml:"pg_db" env-required:"true"`
	SSLMode  string `yaml:"ssl_mode" env-default:"disable"`
	MaxConns int32  `yaml:"max_conns" env-default:"10"`
	MinConns int32  `yaml:"min_conns" env-default:"5"`
	ConnLife int32  `yaml:"conn_life_h" env-default:"1"`
	ConnIdle int32  `yaml:"conn_idle_m" env-default:"1"`
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
