package config

import "os"

const (
	defaultHost = "0.0.0.0"
	defaultPort = "8080"
)

type Config struct {
	Host string
	Port string
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
	return Config{
		Host: host,
		Port: port,
	}, nil
}

func (c Config) Address() string {
	return c.Host + ":" + c.Port
}
