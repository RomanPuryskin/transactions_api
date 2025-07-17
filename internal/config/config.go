package config

import (
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Server struct {
		Port string `env:"SERVER_PORT" envDefault:":8080"`
	}

	Storage struct {
		Host     string `env:"DB_HOST,required"`
		Port     string `env:"DB_PORT,required"`
		User     string `env:"DB_USER,required"`
		Password string `env:"DB_PASSWORD,required"`
		Name     string `env:"DB_NAME,required"`
	}
}

func MustLoad() *Config {

	if err := godotenv.Load("local.env"); err != nil {
		log.Fatal("[MustLoad|load .env file]", err)
	}

	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		log.Fatal("[Mustload|read .env file]", err)
	}

	return cfg
}
