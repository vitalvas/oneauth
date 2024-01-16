package paths

import (
	"fmt"

	"github.com/vitalvas/oneauth/internal/tools"
)

func ServiceFile(name string) (string, error) {
	return tools.InHomeDir("Library", "LaunchAgents", fmt.Sprintf("%s.plist", name))
}
