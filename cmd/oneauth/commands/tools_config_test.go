package commands

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestLoadConfig(t *testing.T) {
	t.Run("FileDoesNotExist", func(t *testing.T) {
		app := cli.NewApp()
		set := flag.NewFlagSet("test", 0)
		set.String("config", "/non/existent/path", "doc")
		c := cli.NewContext(app, set, nil)

		err := loadConfig(c)
		assert.Nil(t, err)
		assert.Nil(t, globalConfig)
	})

	t.Run("FileExists", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config-*.yml")
		if err != nil {
			assert.Error(t, err)
			return
		}
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(`control_socket_path: /tmp/oneauth.sock`)
		if err != nil {
			assert.Error(t, err)
			return
		}
		tmpFile.Close()

		app := cli.NewApp()
		set := flag.NewFlagSet("test", 0)
		set.String("config", tmpFile.Name(), "doc")
		c := cli.NewContext(app, set, nil)

		err = loadConfig(c)
		assert.Nil(t, err)
		assert.NotNil(t, globalConfig)
	})
}
