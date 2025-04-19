package commands

import (
	"errors"
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/cmd/oneauth/service"
)

var serviceEnableCmd = &cli.Command{
	Name:  "enable",
	Usage: "Enable the service",
	Action: func(_ *cli.Context) error {
		if err := service.Install(); err != nil {
			return err
		}

		for i := 0; i < 100; i++ {
			if service.IsRunning() {
				fmt.Println("oneauth agent service has been successfully started")
				return nil
			}

			fmt.Println("waiting for oneauth agent service to start...")
			time.Sleep(500 * time.Millisecond)
		}

		return errors.New("oneauth agent service has been installed, but it is not running")
	},
}
