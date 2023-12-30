package commands

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/buildinfo"
	"github.com/vitalvas/oneauth/internal/tools"
)

func Execute() {
	if tools.IsRoot() {
		log.Fatal("oneauth client must not be run as root")
	}

	app := &cli.App{
		Name:    "oneauth",
		Usage:   "OneAuth is a CLI tool to use unified authentication and authorization",
		Version: buildinfo.Version,
		Commands: []*cli.Command{
			agentCmd,
			infoCmd,
			setupCmd,
			yubikeyCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Println(err)
	}
}
