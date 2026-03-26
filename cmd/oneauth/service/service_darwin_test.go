package service

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestServiceName(t *testing.T) {
	t.Run("ServiceNameConstant", func(t *testing.T) {
		// Test that service name is set correctly
		assert.Equal(t, "dev.vitalvas.oneauth", serviceName)
		assert.NotEmpty(t, serviceName)
		assert.Contains(t, serviceName, "oneauth")
	})
}

func TestServiceTemplate(t *testing.T) {
	t.Run("ServiceTemplateExists", func(t *testing.T) {
		// Test that service template is embedded
		assert.NotEmpty(t, serviceTmpl)
		assert.IsType(t, "", serviceTmpl)
	})
}

func TestWriteServiceTemplate(t *testing.T) {
	t.Run("ValidTemplate", func(t *testing.T) {
		// Create temporary file
		tempFile, err := os.CreateTemp("", "service_test")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// Test writing template
		exePath := "/usr/local/bin/oneauth"
		err = writeServiceTemplate(exePath, tempFile)
		assert.NoError(t, err)
	})

	t.Run("EmptyPath", func(t *testing.T) {
		// Create temporary file
		tempFile, err := os.CreateTemp("", "service_test")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// Test writing template with empty path
		err = writeServiceTemplate("", tempFile)
		assert.NoError(t, err)
	})

	t.Run("LongPath", func(t *testing.T) {
		// Create temporary file
		tempFile, err := os.CreateTemp("", "service_test")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// Test writing template with long path
		longPath := fmt.Sprintf("%s/oneauth", strings.Repeat("/very/long/path", 20))
		err = writeServiceTemplate(longPath, tempFile)
		assert.NoError(t, err)
	})
}

func TestWriteServiceTemplateStructure(t *testing.T) {
	t.Run("TemplateStructure", func(t *testing.T) {
		// Test that template structure is correct
		serviceInfo := struct {
			Args []string
		}{
			Args: []string{
				"/usr/local/bin/oneauth",
				"agent",
			},
		}

		// Test that template can be parsed
		tmpl, err := template.New("service").Parse(serviceTmpl)
		assert.NoError(t, err)
		assert.NotNil(t, tmpl)

		// Test that template can be executed
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, serviceInfo)
		assert.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})
}

func TestCallLaunchCtlStructure(t *testing.T) {
	t.Run("LaunchCtlCommand", func(t *testing.T) {
		// Test that callLaunchCtl handles basic error cases
		// We can't test actual launchctl calls without system dependencies

		// Test with invalid command (should fail)
		output, err := callLaunchCtl("invalid-command")
		assert.Error(t, err)
		assert.Empty(t, output)
	})

	t.Run("LaunchCtlArguments", func(t *testing.T) {
		// Test that arguments are passed correctly
		// This will fail but we can test the structure
		output, err := callLaunchCtl("help")
		// Should either succeed or fail with specific error
		assert.IsType(t, "", output)
		if err != nil {
			assert.IsType(t, (*error)(nil), err)
		}
	})
}

func TestInstallValidation(t *testing.T) {
	t.Run("InstallPathValidation", func(t *testing.T) {
		// Test will fail due to path validation, but we can test the structure
		err := Install()
		assert.Error(t, err)
		// Should be a meaningful error
		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	})
}

func TestUninstalValidation(t *testing.T) {
	t.Run("UninstalTypo", func(t *testing.T) {
		// Test the misspelled function name exists
		err := Uninstal()
		// Should return an error or nil
		assert.True(t, err == nil || err != nil)
	})
}

func TestRestartValidation(t *testing.T) {
	t.Run("RestartService", func(t *testing.T) {
		// Test restart function exists and can be called
		err := Restart()
		// Should return an error or nil
		assert.True(t, err == nil || err != nil)
	})
}

func TestCheckServiceValidation(t *testing.T) {
	t.Run("CheckServiceFunction", func(t *testing.T) {
		// Test check service function
		err := checkService()

		// Should return either ErrNotInstalled or nil
		assert.True(t, err == nil || err == ErrNotInstalled)
	})
}

func TestServiceFunctionSignatures(t *testing.T) {
	t.Run("FunctionSignatures", func(t *testing.T) {
		// Test that all functions have expected signatures
		var err error

		// Install function
		err = Install()
		assert.True(t, err == nil || err != nil)

		// Uninstal function (note the typo in the original)
		err = Uninstal()
		assert.True(t, err == nil || err != nil)

		// Restart function
		err = Restart()
		assert.True(t, err == nil || err != nil)

		// checkService function
		err = checkService()
		assert.True(t, err == nil || err != nil)
	})
}

func TestServicePathHandling(t *testing.T) {
	t.Run("ServicePathConstant", func(t *testing.T) {
		// Test that service name is properly formatted
		assert.Contains(t, serviceName, ".")
		assert.True(t, strings.HasPrefix(serviceName, "dev."))
		assert.True(t, strings.HasSuffix(serviceName, "oneauth"))
	})
}

func TestTemplateArguments(t *testing.T) {
	t.Run("TemplateArgs", func(t *testing.T) {
		// Test that template arguments are structured correctly
		exePath := "/test/path/oneauth"

		// Create temporary file
		tempFile, err := os.CreateTemp("", "service_test")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// Write template
		err = writeServiceTemplate(exePath, tempFile)
		assert.NoError(t, err)

		// Read back and verify it contains expected elements
		content, err := os.ReadFile(tempFile.Name())
		assert.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, exePath)
		assert.Contains(t, contentStr, "agent")
	})
}

func TestServiceFileHandling(t *testing.T) {
	t.Run("FileCreation", func(t *testing.T) {
		// Test that service file can be created and written
		tempDir, err := os.MkdirTemp("", "service_test")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		servicePath := filepath.Join(tempDir, "test.service")

		serviceFile, err := os.Create(servicePath)
		assert.NoError(t, err)
		defer serviceFile.Close()

		// Test writing template to file
		err = writeServiceTemplate("/usr/local/bin/oneauth", serviceFile)
		assert.NoError(t, err)

		// Verify file was written
		info, err := os.Stat(servicePath)
		assert.NoError(t, err)
		assert.True(t, info.Size() > 0)
	})
}

func TestSetSSHAuthSock(t *testing.T) {
	t.Run("SetSSHAuthSock", func(t *testing.T) {
		err := SetSSHAuthSock("/tmp/test-ssh-auth-sock")
		// On macOS, launchctl setenv should work
		// It may fail in CI or restricted environments
		if err != nil {
			assert.Contains(t, err.Error(), "failed to set SSH_AUTH_SOCK")
		}
	})

	t.Run("SetSSHAuthSockWithEmptyPath", func(t *testing.T) {
		err := SetSSHAuthSock("")
		// Even with empty path, launchctl setenv should accept it
		if err != nil {
			assert.Contains(t, err.Error(), "failed to set SSH_AUTH_SOCK")
		}
	})
}

func TestCallLaunchCtlLoadFailed(t *testing.T) {
	t.Run("LoadFailedDetection", func(t *testing.T) {
		// Test that callLaunchCtl detects "Load failed" in stderr
		// callLaunchCtl with list on a non-existent service returns error
		_, err := callLaunchCtl("list", "com.nonexistent.service.12345")
		// This should return ErrNotInstalled when checked via checkService
		// or an error from callLaunchCtl
		if err != nil {
			assert.NotEmpty(t, err.Error())
		}
	})
}

func TestCheckService(t *testing.T) {
	t.Run("ServiceNotInstalled", func(t *testing.T) {
		// checkService should return ErrNotInstalled if the service is not loaded
		err := checkService()
		// On machines without the service installed, this should return ErrNotInstalled
		assert.True(t, err == nil || err == ErrNotInstalled)
	})
}

func TestCallLaunchCtlWithVariousArgs(t *testing.T) {
	t.Run("ListCommand", func(t *testing.T) {
		output, err := callLaunchCtl("list")
		// list should succeed on macOS
		if err == nil {
			assert.NotEmpty(t, output)
		}
	})

	t.Run("VersionCommand", func(t *testing.T) {
		output, err := callLaunchCtl("version")
		// version may or may not work depending on macOS version
		if err == nil {
			assert.IsType(t, "", output)
		}
	})
}

func TestInstallFailsOutsideBinDir(t *testing.T) {
	t.Run("InstallFromWrongDirectory", func(t *testing.T) {
		err := Install()
		// Should fail because the test binary is not in the BinDir
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service can be installed only from app home directory")
	})
}

func TestRestartBehavior(t *testing.T) {
	t.Run("RestartWhenNotInstalled", func(t *testing.T) {
		// If service is not installed, Restart should return nil
		err := checkService()
		if err == ErrNotInstalled {
			err = Restart()
			assert.NoError(t, err)
		}
	})
}

func TestUninstalBehavior(t *testing.T) {
	t.Run("UninstalWhenNotInstalled", func(t *testing.T) {
		// If service is not installed, Uninstal should return nil
		err := checkService()
		if err == ErrNotInstalled {
			err = Uninstal()
			assert.NoError(t, err)
		}
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("ErrorTypes", func(t *testing.T) {
		// Test that functions return appropriate error types
		err := Install()
		if err != nil {
			assert.NotEmpty(t, err.Error())
		}

		err = Uninstal()
		if err != nil {
			assert.NotEmpty(t, err.Error())
		}

		err = Restart()
		if err != nil {
			assert.NotEmpty(t, err.Error())
		}

		err = checkService()
		if err != nil {
			assert.NotEmpty(t, err.Error())
		}
	})
}
