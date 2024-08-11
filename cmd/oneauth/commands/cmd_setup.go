package commands

import "github.com/urfave/cli/v2"

var setupCmd = &cli.Command{
	Name:  "setup",
	Usage: "Setup a YubiKey",
	Subcommands: []*cli.Command{
		setupNewCmd,
		setupPivSlotCmd,
	},
}
