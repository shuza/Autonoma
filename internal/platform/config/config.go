package config

import (
	"fmt"
	"os"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = "8080"
)

type Config struct {
	Host        string
	Port        string
	DatabaseURL string
}

func Load() Config {
	cfg, err := LoadFromEnv()
	if err != nil {
		panic(err)
	}

	return cfg
}

func LoadFromEnv() (Config, error) {
	host := os.Getenv("AUTONOMA_API_HOST")
	if host == "" {
		host = defaultHost
	}

	port := os.Getenv("AUTONOMA_API_PORT")
	if port == "" {
		port = defaultPort
	}

	if port == "0" {
		return Config{}, fmt.Errorf("AUTONOMA_API_PORT must not be 0")
	}

	fmt.Println("====  database : ", os.Getenv("AUTONOMA_DATABASE_URL"))

	return Config{
		Host:        host,
		Port:        port,
		DatabaseURL: os.Getenv("AUTONOMA_DATABASE_URL"),
	}, nil
}

func (c Config) Address() string {
	return c.Host + ":" + c.Port
}
