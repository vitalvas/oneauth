package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/cmd/oneauth/service"
)

var serviceRestartCmd = &cli.Command{
	Name:  "restart",
	Usage: "Restart the service",
	Action: func(c *cli.Context) error {
		if err := service.ServiceRestart(); err != nil {
			return err
		}

		fmt.Println("done...")

		return nil
	},
}
