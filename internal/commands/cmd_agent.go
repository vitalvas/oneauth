package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/gokit/xcmd"
	"github.com/vitalvas/oneauth/internal/sshagent"
	"github.com/vitalvas/oneauth/internal/tools"
	"golang.org/x/sync/errgroup"
)

var agentCmd = &cli.Command{
	Name:  "agent",
	Usage: "SSH Agent",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "socket",
			Value: filepath.Join(tools.GetHomeDir(), ".oneauth/ssh-agent.sock"),
		},
		&cli.Uint64Flag{
			Name:  "serial",
			Usage: "YubiKey serial number",
		},
	},
	Before: selectYubiKey,
	Action: func(c *cli.Context) error {
		serial := c.Uint64("serial")
		if serial == 0 {
			return fmt.Errorf("serial is required")
		}

		socketPath := c.String("socket")
		if _, err := os.Stat(socketPath); err == nil {
			os.Remove(socketPath)
		}

		if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		agent, err := sshagent.New(uint32(serial))
		if err != nil {
			return fmt.Errorf("failed to create agent: %w", err)
		}

		group, ctx := errgroup.WithContext(c.Context)

		group.Go(func() error {
			return agent.ListenAndServe(ctx, socketPath)
		})

		group.Go(func() error {
			err := xcmd.WaitInterrupted(ctx)
			log.Println("shutting down agent")
			agent.Shutdown()
			return err
		})

		return group.Wait()
	},
}
