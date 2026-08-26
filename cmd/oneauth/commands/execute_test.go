package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/buildinfo"
)

func TestNewApp(t *testing.T) {
	app := newApp("/mock/config/path")

	t.Run("Metadata", func(t *testing.T) {
		assert.Equal(t, "oneauth", app.Name)
		assert.Equal(t, "OneAuth is a CLI tool to use unified authentication and authorization", app.Usage)
		assert.Equal(t, "Details: https://oneauth.vitalvas.dev", app.Description)
		assert.Equal(t, buildinfo.Version, app.Version)
	})

	t.Run("ConfigFlag", func(t *testing.T) {
		require.Len(t, app.Flags, 1)

		configFlag, ok := app.Flags[0].(*cli.PathFlag)
		require.True(t, ok)
		assert.Equal(t, "config", configFlag.Name)
		assert.Equal(t, "path to config file", configFlag.Usage)
		assert.Equal(t, "/mock/config/path", configFlag.Value)
	})

	t.Run("Commands", func(t *testing.T) {
		require.Len(t, app.Commands, 6)

		names := make([]string, 0, len(app.Commands))
		for _, cmd := range app.Commands {
			names = append(names, cmd.Name)
		}
		assert.Contains(t, names, "agent")
		assert.Contains(t, names, "info")
		assert.Contains(t, names, "setup")
		assert.Contains(t, names, "service")
		assert.Contains(t, names, "yubikey")
		assert.Contains(t, names, "update")
	})
}

func TestNewAppRun(t *testing.T) {
	t.Run("Help", func(t *testing.T) {
		app := newApp("/mock/config/path")
		err := app.Run([]string{"oneauth", "--help"})
		assert.NoError(t, err)
	})

	t.Run("UnknownFlag", func(t *testing.T) {
		app := newApp("/mock/config/path")
		err := app.Run([]string{"oneauth", "--does-not-exist"})
		assert.Error(t, err)
	})

	t.Run("ConfigFlagPassthrough", func(t *testing.T) {
		app := newApp("/mock/config/path")
		err := app.Run([]string{"oneauth", "--config", "/tmp/custom.yaml", "--help"})
		assert.NoError(t, err)
	})
}

func TestCommandMetadata(t *testing.T) {
	t.Run("ServiceCmd", func(t *testing.T) {
		assert.Equal(t, "service", serviceCmd.Name)
		assert.Equal(t, "Service management", serviceCmd.Usage)
		require.Len(t, serviceCmd.Subcommands, 3)

		subNames := make([]string, 0, len(serviceCmd.Subcommands))
		for _, sub := range serviceCmd.Subcommands {
			subNames = append(subNames, sub.Name)
		}
		assert.Contains(t, subNames, "enable")
		assert.Contains(t, subNames, "disable")
		assert.Contains(t, subNames, "restart")
	})

	t.Run("InfoCmd", func(t *testing.T) {
		assert.Equal(t, "info", infoCmd.Name)
		assert.Equal(t, "Prints detailed information", infoCmd.Usage)
		assert.NotNil(t, infoCmd.Action)
	})

	t.Run("UpdateCmd", func(t *testing.T) {
		assert.Equal(t, "update", updateCmd.Name)
		assert.Equal(t, "update oneauth", updateCmd.Usage)
		assert.NotNil(t, updateCmd.Action)
	})

	t.Run("SetupCmd", func(t *testing.T) {
		assert.Equal(t, "setup", setupCmd.Name)
		assert.Equal(t, "Setup a YubiKey", setupCmd.Usage)
		assert.NotEmpty(t, setupCmd.Subcommands)
	})

	t.Run("YubikeyCmd", func(t *testing.T) {
		assert.Equal(t, "yubikey", yubikeyCmd.Name)
		assert.NotNil(t, yubikeyCmd)
	})

	t.Run("ServiceSubcommands", func(t *testing.T) {
		assert.Equal(t, "enable", serviceEnableCmd.Name)
		assert.Equal(t, "Enable the service", serviceEnableCmd.Usage)
		assert.NotNil(t, serviceEnableCmd.Action)

		assert.Equal(t, "disable", serviceDisableCmd.Name)
		assert.Equal(t, "Disable the service", serviceDisableCmd.Usage)
		assert.NotNil(t, serviceDisableCmd.Action)

		assert.Equal(t, "restart", serviceRestartCmd.Name)
		assert.Equal(t, "Restart the service", serviceRestartCmd.Usage)
		assert.NotNil(t, serviceRestartCmd.Action)
	})
}

func TestInfoKeyStruct(t *testing.T) {
	t.Run("InfoKeyFields", func(t *testing.T) {
		key := InfoKey{
			Name:    "test-key",
			Serial:  "12345",
			Version: "1.0.0",
		}

		assert.Equal(t, "test-key", key.Name)
		assert.Equal(t, "12345", key.Serial)
		assert.Equal(t, "1.0.0", key.Version)
	})

	t.Run("InfoDataStructure", func(t *testing.T) {
		data := infoData{
			Keys: []InfoKey{
				{Name: "key1", Serial: "1", Version: "1.0"},
				{Name: "key2", Serial: "2", Version: "2.0"},
			},
		}

		require.Len(t, data.Keys, 2)
		assert.Equal(t, "key1", data.Keys[0].Name)
		assert.Equal(t, "key2", data.Keys[1].Name)
	})
}

func TestInfoTemplate(t *testing.T) {
	t.Run("TemplateValid", func(t *testing.T) {
		assert.NotEmpty(t, infoTmpl)
		assert.Contains(t, infoTmpl, "Keys")
	})
}
