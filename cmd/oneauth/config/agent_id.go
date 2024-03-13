package config

import (
	"os"

	"github.com/google/uuid"
	"github.com/vitalvas/oneauth/cmd/oneauth/paths"
)

func LoadOrCreateAgentID() (uuid.UUID, error) {
	agentIDPath, err := paths.AgentID()
	if err != nil {
		return [16]byte{}, err
	}

	if stat, err := os.Stat(agentIDPath); err == nil && !stat.IsDir() {
		file, err2 := os.ReadFile(agentIDPath)
		if err2 != nil {
			return [16]byte{}, err2
		}

		if id, err3 := uuid.Parse(string(file)); err3 == nil {
			return id, nil
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return [16]byte{}, err
	}

	if err := os.WriteFile(agentIDPath, []byte(id.String()), 0600); err != nil {
		return [16]byte{}, err
	}

	return id, nil
}
