package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vitalvas/oneauth/cmd/oneauth/paths"
	"github.com/vitalvas/oneauth/internal/tools"
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

	if err := loadYamlFile(filePath, conf); err != nil {
		return nil, err
	}

	// Expand ~ in agent socket paths
	if err := expandAgentPaths(conf); err != nil {
		return nil, err
	}

	agentID, err := LoadOrCreateAgentID()
	if err != nil {
		return nil, err
	}

	conf.AgentID = agentID

	return conf, nil
}

func expandAgentPaths(conf *Config) error {
	homeDir, err := tools.GetHomeDir()
	if err != nil {
		return err
	}

	for name, agent := range conf.Agents {
		// Set default socket path if not specified
		if agent.SocketPath == "" {
			defaultPath, err := paths.NamedAgentSocket(name)
			if err != nil {
				return fmt.Errorf("failed to get default socket path for agent %s: %w", name, err)
			}
			agent.SocketPath = defaultPath
		} else if strings.HasPrefix(agent.SocketPath, "~/") {
			// Expand ~ in socket paths
			agent.SocketPath = homeDir + agent.SocketPath[1:]
		}
		conf.Agents[name] = agent
	}

	return nil
}

func loadYamlFile(filePath string, v *Config) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true) // fail on unknown fields

	return decoder.Decode(v)
}
