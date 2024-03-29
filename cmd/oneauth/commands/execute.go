package commands

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/cmd/oneauth/paths"
	"github.com/vitalvas/oneauth/internal/buildinfo"
	"github.com/vitalvas/oneauth/internal/tools"
)

func Execute() {
	if tools.IsRoot() {
		log.Fatal("oneauth client must not be run as root")
	}

	configPath, err := paths.Config()
	if err != nil {
		log.Fatal(err)
	}

	app := &cli.App{
		Name:        "oneauth",
		Usage:       "OneAuth is a CLI tool to use unified authentication and authorization",
		Description: "Details: https://oneauth.vitalvas.dev",
		Version:     buildinfo.Version,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:  "config",
				Usage: "path to config file",
				Value: configPath,
			},
		},
		Commands: []*cli.Command{
			agentCmd,
			infoCmd,
			setupCmd,
			serviceCmd,
			yubikeyCmd,
			updateCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Println(err)
	}
}
