package config

import (
	"os"

	"github.com/vitalvas/oneauth/cmd/oneauth/paths"
	"gopkg.in/yaml.v3"
)

func Load(filePath string) (*Config, error) {
	socketPath, err := paths.AgentSocket()
	if err != nil {
		return nil, err
	}

	conf := &Config{
		Socket: Socket{
			Type: "unix",
			Path: socketPath,
		},
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(&conf); err != nil {
		return nil, err
	}

	return conf, nil
}
