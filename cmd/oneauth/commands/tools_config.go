package commands

import (
	"os"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/cmd/oneauth/config"
)

var globalConfig *config.Config

func loadConfig(c *cli.Context) error {
	configPath := c.String("config")

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	var err error

	globalConfig, err = config.Load(configPath)

	return err
}
