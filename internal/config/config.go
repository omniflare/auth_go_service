package config

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

// new env vars so here.
// any other config related things go in this package ;

type Config struct {
	PORT        string `env:"PORT,required"`
	Database    string `env:"DATABASE_URL,required"`
	FirebaseAPI string `env:"FIREBASE_API,required"`
}

func NewEnvConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error while accessing env file: %s", err.Error())
	}
	config := &Config{}
	if err := env.Parse(config); err != nil {
		log.Fatalf("Error parsing env config %e", err)
	}
	return config
}
