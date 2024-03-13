package paths

import "github.com/vitalvas/oneauth/internal/tools"

const oneauthDir = ".oneauth"

func AgentID() (string, error) {
	return tools.InHomeDir(oneauthDir, "agent_id")
}

func AgentSocket() (string, error) {
	return tools.InHomeDir(oneauthDir, "ssh-agent.sock")
}

func ControlSocket() (string, error) {
	return tools.InHomeDir(oneauthDir, "control.sock")
}

func Config() (string, error) {
	return tools.InHomeDir(oneauthDir, "config.yaml")
}

func BinDir() (string, error) {
	return tools.InHomeDir(oneauthDir, "bin")
}

func LogDir() (string, error) {
	return tools.InHomeDir(oneauthDir, "log")
}
