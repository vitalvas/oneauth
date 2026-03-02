package paths

import (
	"fmt"

	"github.com/vitalvas/oneauth/internal/tools"
)

const oneauthDir = ".oneauth"

func AgentID() (string, error) {
	return tools.InHomeDir(oneauthDir, "agent_id")
}

func AgentSocket() (string, error) {
	return tools.InHomeDir(oneauthDir, "ssh-agent.sock")
}

// NamedAgentSocket returns the default socket path for a named soft-key agent
func NamedAgentSocket(name string) (string, error) {
	return tools.InHomeDir(oneauthDir, fmt.Sprintf("ssh-agent-%s.sock", name))
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
