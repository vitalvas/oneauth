package server

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/buildinfo"
	"github.com/vitalvas/oneauth/internal/yubico"
)

type Server struct {
	config *Config

	yubico *yubico.YubiAuth
}

func Execute() {
	srv := Server{}

	app := &cli.App{
		Name:    "oneauth-server",
		Version: buildinfo.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Value:   "config.json",
				EnvVars: []string{"ONEAUTH_CONFIG_FILE"},
			},
		},
		Before: srv.loadConfig,
		Action: srv.runHTTPServer,
	}

	if err := app.Run(os.Args); err != nil {
		log.Println(err)
	}
}
