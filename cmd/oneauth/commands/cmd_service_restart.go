package commands

import (
	"errors"
	"fmt"
	"time"

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

		for i := 0; i < 100; i++ {
			if service.IsRunning() {
				fmt.Println("oneauth agent service has been successfully restarted")
				return nil
			}

			fmt.Println("waiting for oneauth agent service to start...")
			time.Sleep(500 * time.Millisecond)
		}

		return errors.New("oneauth agent service failed to restart")
	},
}
