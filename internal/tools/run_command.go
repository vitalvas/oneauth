package tools

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func RunCommand(cmdName string, env map[string]string) error {
	cmdSlice := strings.Split(cmdName, " ")
	cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...) //nolint:gosec

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Stdout = nil

	var stderrBuffer bytes.Buffer
	cmd.Stderr = &stderrBuffer

	if err := cmd.Run(); err != nil {
		return err
	}

	if stderrBuffer.Len() > 0 {
		return errors.New(stderrBuffer.String())
	}

	return nil
}
