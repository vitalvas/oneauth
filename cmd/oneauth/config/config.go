package config

import (
	"fmt"
	"os"
	"time"

	"github.com/vitalvas/oneauth/cmd/oneauth/paths"
	"gopkg.in/yaml.v3"
)

func Load(filePath string) (*Config, error) {
	agentSocketPath, err := paths.AgentSocket()
	if err != nil {
		return nil, err
	}

	controlSocketPath, err := paths.ControlSocket()
	if err != nil {
		return nil, err
	}

	agentLogDir, err := paths.LogDir()
	if err != nil {
		return nil, err
	}

	conf := &Config{
		ControlSocketPath: controlSocketPath,
		AgentLogPath:      fmt.Sprintf("%s/agent_%d.log", agentLogDir, time.Now().Year()),
		Socket: Socket{
			Type: "unix",
			Path: agentSocketPath,
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

	agentID, err := LoadOrCreateAgentID()
	if err != nil {
		return nil, err
	}

	conf.AgentID = agentID

	return conf, nil
}
