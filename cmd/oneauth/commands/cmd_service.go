package commands

import "github.com/urfave/cli/v2"

var serviceCmd = &cli.Command{
	Name:        "service",
	Usage:       "Service management",
	Description: "Manage OneAuth Agent service",
	Subcommands: []*cli.Command{
		serviceEnableCmd,
		serviceDisableCmd,
		serviceRestartCmd,
	},
}
