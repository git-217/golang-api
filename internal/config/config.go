package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string      `yaml:"env" env-default:"local"`
	PsqlConn    pgParams    `yaml:"psql_params"`
	HttpServer  HTTPServer  `yaml:"http_server"`
	RateLimiter RateLimiter `yaml:"rate_limiter"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	User        string        `yaml:"user" env-required:"true"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
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

type RateLimiter struct {
	MaxTokens  float64 `yaml:"max_tokens" env-default:"30"`
	RefillRate float64 `yaml:"refill_rate" env-default:"15"`
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
		log.Fatalf("Can't read config file: %s.", configPath)
	}

	return &cfg
}
