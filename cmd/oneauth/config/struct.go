package config

import "github.com/google/uuid"

type Config struct {
	AgentID uuid.UUID `yaml:"-"`

	ControlSocketPath string  `yaml:"control_socket_path,omitempty"`
	AgentLogPath      string  `yaml:"agent_log_path,omitempty"`
	Socket            Socket  `yaml:"socket,omitempty"`
	Keyring           Keyring `yaml:"keyring,omitempty"`

	// Agents defines additional soft-key-only SSH agents
	Agents map[string]AgentConfig `yaml:"agents,omitempty"`
}

// AgentConfig defines configuration for an additional soft-key SSH agent
type AgentConfig struct {
	SocketPath     string `yaml:"socket_path"`
	KeepKeySeconds int64  `yaml:"keep_key_seconds,omitempty"`
}

type Socket struct {
	Type string `yaml:"type,omitempty"`
	Path string `yaml:"path,omitempty"`
}

type Keyring struct {
	Yubikey        KeyringYubikey `yaml:"yubikey,omitempty"`
	BeforeSignHook string         `yaml:"before_sign_hook,omitempty"`
	KeepKeySeconds int64          `yaml:"keep_key_seconds,omitempty"`
}

type KeyringYubikey struct {
	Serial uint32 `yaml:"serial,omitempty"`
}
