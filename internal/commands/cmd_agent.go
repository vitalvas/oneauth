package commands

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/sshagent"
)

var agentCmd = &cli.Command{
	Name:  "agent",
	Usage: "SSH Agent",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "socket",
			Value: filepath.Join(os.Getenv("HOME"), ".oneauth/ssh-agent.sock"),
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

		listener, err := net.Listen("unix", socketPath)
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}

		for {
			conn, err := listener.Accept()
			if err != nil {
				if err, ok := err.(sshagent.Temporary); ok && err.Temporary() {
					log.Printf("temporary accept error: %v", err)
					time.Sleep(time.Second)
					continue
				}

				return fmt.Errorf("failed to accept: %w", err)
			}

			go agent.HandleConn(conn)
		}
	},
}
