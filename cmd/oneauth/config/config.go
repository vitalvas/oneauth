package config

import (
	"fmt"
	"os"
	"time"

	"github.com/vitalvas/oneauth/cmd/oneauth/paths"
	"github.com/vitalvas/oneauth/internal/tools"
	"gopkg.in/yaml.v3"
)

func Load(filePath string) (*Config, error) {
	rootDir, err := paths.RootDir()
	if err != nil {
		return nil, err
	}

	if err := tools.MkDir(rootDir, 0700); err != nil {
		return nil, err
	}

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

	agentID, err := LoadOrCreateAgentID()
	if err != nil {
		return nil, err
	}

	conf.AgentID = agentID

	return conf, nil
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
