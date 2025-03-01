package config

type Config struct {
	GrpcServerAddr string
}

func NewConfig() *Config {
	return &Config{GrpcServerAddr: ":8888"}
}
