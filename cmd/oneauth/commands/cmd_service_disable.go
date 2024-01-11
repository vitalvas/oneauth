package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/cmd/oneauth/service"
)

var serviceDisableCmd = &cli.Command{
	Name:  "disable",
	Usage: "Disable the service",
	Action: func(c *cli.Context) error {
		if err := service.Uninstal(); err != nil {
			return err
		}

		fmt.Println("oneauth agent service disabled")

		return nil
	},
}
