package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/cmd/oneauth/service"
)

var serviceRestartCmd = &cli.Command{
	Name:  "restart",
	Usage: "Restart the service",
	Action: func(_ *cli.Context) error {
		if err := service.Restart(); err != nil {
			return err
		}

		fmt.Println("done...")

		return nil
	},
}
