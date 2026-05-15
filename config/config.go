package config

const (
	DefaultServerAddress = "0.0.0.0:8080"
	DefaultClientServer  = "http://localhost:8080"
)

type Config struct {
	ServerAddress string
}

func NewConfig() *Config {
	return &Config{
		ServerAddress: DefaultServerAddress,
	}
}
