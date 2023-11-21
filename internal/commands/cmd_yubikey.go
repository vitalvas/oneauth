package commands

import (
	"github.com/urfave/cli/v2"
)

var yubikeyCmd = &cli.Command{
	Name:  "yubikey",
	Usage: "Yubikey related commands",
	Subcommands: []*cli.Command{
		yubikeyListCmd,
		yubikeyResetCmd,
	},
}
