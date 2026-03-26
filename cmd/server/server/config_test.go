package server

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestConfigStructs(t *testing.T) {
	t.Run("ConfigJSONTags", func(t *testing.T) {
		cfg := Config{
			Yubico: ConfigYubico{
				ClientID:     123,
				ClientSecret: "test-secret",
			},
		}

		data, err := json.Marshal(cfg)
		require.NoError(t, err)

		var parsed Config
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)
		assert.Equal(t, 123, parsed.Yubico.ClientID)
		assert.Equal(t, "test-secret", parsed.Yubico.ClientSecret)
	})

	t.Run("ConfigYubicoJSONTags", func(t *testing.T) {
		yc := ConfigYubico{
			ClientID:     42,
			ClientSecret: "secret123",
		}

		data, err := json.Marshal(yc)
		require.NoError(t, err)

		assert.Contains(t, string(data), "client_id")
		assert.Contains(t, string(data), "client_secret")
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("ValidConfigFile", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		cfg := Config{
			Yubico: ConfigYubico{
				ClientID:     999,
				ClientSecret: "my-secret",
			},
		}

		data, err := json.Marshal(cfg)
		require.NoError(t, err)
		_, err = tmpFile.Write(data)
		require.NoError(t, err)
		tmpFile.Close()

		srv := &Server{config: &Config{}}

		// We cannot easily create a cli.Context, but we can test the JSON decode logic
		file, err := os.Open(tmpFile.Name())
		require.NoError(t, err)
		defer file.Close()

		err = json.NewDecoder(file).Decode(srv.config)
		require.NoError(t, err)
		assert.Equal(t, 999, srv.config.Yubico.ClientID)
		assert.Equal(t, "my-secret", srv.config.Yubico.ClientSecret)
	})

	t.Run("InvalidConfigFile", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("not valid json")
		require.NoError(t, err)
		tmpFile.Close()

		var cfg Config
		file, err := os.Open(tmpFile.Name())
		require.NoError(t, err)
		defer file.Close()

		err = json.NewDecoder(file).Decode(&cfg)
		assert.Error(t, err)
	})

	t.Run("MissingConfigFile", func(t *testing.T) {
		_, err := os.Open("/nonexistent/config.json")
		assert.Error(t, err)
	})
}

func TestLoadConfigViaCLI(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		cfg := Config{
			Yubico: ConfigYubico{
				ClientID:     42,
				ClientSecret: "c2VjcmV0",
			},
		}
		data, err := json.Marshal(cfg)
		require.NoError(t, err)
		_, err = tmpFile.Write(data)
		require.NoError(t, err)
		tmpFile.Close()

		srv := &Server{}
		app := &cli.App{
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "config",
					Value: tmpFile.Name(),
				},
			},
			Action: func(c *cli.Context) error {
				return srv.loadConfig(c)
			},
		}

		err = app.Run([]string{"app"})
		require.NoError(t, err)
		assert.NotNil(t, srv.config)
		assert.Equal(t, 42, srv.config.Yubico.ClientID)
		assert.Equal(t, "c2VjcmV0", srv.config.Yubico.ClientSecret)
	})

	t.Run("MissingConfigFile", func(t *testing.T) {
		srv := &Server{}
		app := &cli.App{
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "config",
					Value: "/nonexistent/config.json",
				},
			},
			Action: func(c *cli.Context) error {
				return srv.loadConfig(c)
			},
		}

		err := app.Run([]string{"app"})
		assert.Error(t, err)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("{invalid json}")
		require.NoError(t, err)
		tmpFile.Close()

		srv := &Server{}
		app := &cli.App{
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "config",
					Value: tmpFile.Name(),
				},
			},
			Action: func(c *cli.Context) error {
				return srv.loadConfig(c)
			},
		}

		err = app.Run([]string{"app"})
		assert.Error(t, err)
	})

	t.Run("EmptyConfig", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("{}")
		require.NoError(t, err)
		tmpFile.Close()

		srv := &Server{}
		app := &cli.App{
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "config",
					Value: tmpFile.Name(),
				},
			},
			Action: func(c *cli.Context) error {
				return srv.loadConfig(c)
			},
		}

		err = app.Run([]string{"app"})
		require.NoError(t, err)
		assert.NotNil(t, srv.config)
		assert.Equal(t, 0, srv.config.Yubico.ClientID)
	})
}
