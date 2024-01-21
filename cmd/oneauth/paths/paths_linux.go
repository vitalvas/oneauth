package paths

import (
	"fmt"

	"github.com/vitalvas/oneauth/internal/tools"
)

func ServiceFile(name string) (string, error) {
	return tools.InHomeDir(".config", "systemd", "user", fmt.Sprintf("%s.service", name))
}
