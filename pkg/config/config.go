package config

import (
	"github.com/joeshaw/envdecode"
	pkgerrors "github.com/pkg/errors"
)

type Config struct {
	RPCEndpoint string `env:"RPC_ENDPOINT,required"`
	ServerHost  string `env:"SERVER_HOST,default:0.0.0.0"`
	ServerPort  int    `env:"SERVER_PORT,default:8080"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := envdecode.Decode(&cfg); err != nil {
		return nil, pkgerrors.Wrap(err, "decode env")
	}

	return &cfg, nil
}
