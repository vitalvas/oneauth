package commands

import "github.com/urfave/cli/v2"

var dataCmd = &cli.Command{
	Name:  "data",
	Usage: "Data related commands",
	Subcommands: []*cli.Command{
		dataSignCmd,
	},
}
