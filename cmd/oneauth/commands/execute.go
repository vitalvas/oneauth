package commands

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/tools"
)

func Execute() {
	if tools.IsRoot() {
		log.Fatal("oneauth client must not be run as root")
	}

	version := fmt.Sprintf("0.0.%d", time.Now().Unix())

	app := &cli.App{
		Name:    "oneauth",
		Usage:   "OneAuth is a CLI tool to use unified authentication and authorization",
		Version: version,
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
