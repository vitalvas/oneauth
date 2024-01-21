package config

type Config struct {
	ControlSocketPath string  `yaml:"control_socket_path,omitempty"`
	AgentLogPath      string  `yaml:"agent_log_path,omitempty"`
	Socket            Socket  `yaml:"socket,omitempty"`
	Keyring           Keyring `yaml:"keyring,omitempty"`
}

type Socket struct {
	Type string `yaml:"type,omitempty"`
	Path string `yaml:"path,omitempty"`
}

type Keyring struct {
	Yubikey KeyringYubikey `yaml:"yubikey,omitempty"`
}

type KeyringYubikey struct {
	Serial uint32 `yaml:"serial,omitempty"`
}
