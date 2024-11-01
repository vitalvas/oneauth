package service

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/vitalvas/oneauth/cmd/oneauth/paths"
)

const serviceName = "dev.vitalvas.oneauth"

//go:embed template/launchd.service
var serviceTmpl string

func Install() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	appHomeDir, err := paths.BinDir()
	if err != nil {
		return fmt.Errorf("failed to get app home directory: %w", err)
	}

	if !strings.HasPrefix(execPath, appHomeDir) {
		return errors.New("service can be installed only from app home directory")
	}

	servicePath, err := paths.ServiceFile(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service path: %w", err)
	}

	servicePathDir := filepath.Dir(servicePath)

	if _, err := os.Stat(servicePathDir); err != nil {
		if err := os.MkdirAll(servicePathDir, 0700); err != nil {
			return fmt.Errorf("failed to create service directory: %w", err)
		}
	}

	if _, err := os.Stat(servicePath); err == nil {
		if _, err := callLaunchCtl("unload", servicePath); err != nil {
			return fmt.Errorf("failed to unload service: %w", err)
		}
	}

	serviceFile, err := os.Create(servicePath)
	if err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
	}

	defer serviceFile.Close()

	if err := writeServiceTemplate(execPath, serviceFile); err != nil {
		return fmt.Errorf("failed to write service template: %w", err)
	}

	if _, err := callLaunchCtl("load", "-w", servicePath); err != nil {
		return fmt.Errorf("failed to load service: %w", err)
	}

	return nil
}

func Uninstal() error {
	if err := checkService(); err == ErrNotInstalled {
		return nil
	}

	servicePath, err := paths.ServiceFile(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service path: %w", err)
	}

	if _, err := callLaunchCtl("unload", servicePath); err != nil {
		return fmt.Errorf("failed to unload service: %w", err)
	}

	if err := os.RemoveAll(servicePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to remove service file: %w", err)
	}

	return nil
}

func Restart() error {
	if err := checkService(); err == ErrNotInstalled {
		return nil
	}

	if _, err := callLaunchCtl("stop", serviceName); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	return nil
}

func checkService() error {
	if _, err := callLaunchCtl("list", serviceName); err != nil {
		return ErrNotInstalled
	}

	return nil
}

func callLaunchCtl(args ...string) (string, error) {
	c := exec.Command("/bin/launchctl", args...)

	var (
		stdout, stdin, stderr bytes.Buffer
	)

	c.Stdin = &stdin
	c.Stdout = &stdout
	c.Stderr = &stderr

	if err := c.Run(); err != nil {
		stderr.WriteTo(os.Stderr)
		return "", fmt.Errorf("failed to run launchctl: %w", err)
	}

	strErr := stderr.String()
	os.Stderr.WriteString(strErr)

	// oh, apple...
	if strings.Contains(strErr, "Load failed") {
		return "", errors.New("launchctl: load failed")
	}

	return stdout.String(), nil
}

func writeServiceTemplate(exePath string, serviceFile *os.File) error {
	serviceInfo := struct {
		Args []string
	}{
		Args: []string{
			exePath,
			"agent",
		},
	}

	return template.Must(
		template.New("service").Parse(serviceTmpl),
	).Execute(serviceFile, serviceInfo)
}

func IsRunning() bool {
	output, err := callLaunchCtl("list", serviceName)
	if err != nil {
		return false
	}

	if !strings.Contains(output, "PID") {
		return false
	}

	return true
}
