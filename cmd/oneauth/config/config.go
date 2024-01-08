package config

import (
	"os"
	"path/filepath"

	"github.com/vitalvas/oneauth/internal/tools"
	"gopkg.in/yaml.v3"
)

func Load(filePath string) (*Config, error) {
	conf := &Config{
		Socket: Socket{
			Type: "unix",
			Path: filepath.Join(tools.GetHomeDir(), ".oneauth/ssh-agent.sock"),
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
