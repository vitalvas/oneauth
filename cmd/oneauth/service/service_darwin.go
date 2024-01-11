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

	"github.com/vitalvas/oneauth/internal/tools"
)

const serviceName = "dev.vitalvas.oneauth"

//go:embed template/launchd.service
var serviceTmpl string

func ServiceInstall() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	appHomeDir, err := tools.InHomeDir(".oneauth", "bin")
	if err != nil {
		return fmt.Errorf("failed to get app home directory: %w", err)
	}

	if !strings.HasPrefix(execPath, appHomeDir) {
		return errors.New("service can be installed only from app home directory")
	}

	servicePath, err := getServicePath()
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

func ServiceUninstal() error {
	if err := checkService(); err == ErrNotInstalled {
		return nil
	}

	servicePath, err := getServicePath()
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

func ServiceRestart() error {
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

func getServicePath() (string, error) {
	return tools.InHomeDir("Library", "LaunchAgents", fmt.Sprintf("%s.plist", serviceName))
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
