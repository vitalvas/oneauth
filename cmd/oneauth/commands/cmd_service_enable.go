package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/cmd/oneauth/service"
)

var serviceEnableCmd = &cli.Command{
	Name:  "enable",
	Usage: "Enable the service",
	Action: func(c *cli.Context) error {
		if err := service.Install(); err != nil {
			return err
		}

		fmt.Println("oneauth agent service has been successfully installed and probably started")

		return nil
	},
}
