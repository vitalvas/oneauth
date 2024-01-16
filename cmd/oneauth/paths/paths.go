package paths

import "github.com/vitalvas/oneauth/internal/tools"

const oneauthDir = ".oneauth"

func AgentSocket() (string, error) {
	return tools.InHomeDir(oneauthDir, "ssh-agent.sock")
}

func Config() (string, error) {
	return tools.InHomeDir(oneauthDir, "config.yaml")
}

func BinDir() (string, error) {
	return tools.InHomeDir(oneauthDir, "bin")
}
