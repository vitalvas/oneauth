package server

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/buildinfo"
	"golang.org/x/crypto/ssh"
)

type Server struct {
	serverURL *url.URL
	sshConfig *ssh.ServerConfig
}

func Execute() {
	srv := &Server{}

	app := &cli.App{
		Name:    "oneauth-ssh-test-server",
		Version: buildinfo.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "server-url",
				Usage:   "OneAuth server URL",
				Value:   "http://127.0.0.1:8080",
				EnvVars: []string{"ONEAUTH_SERVER_URL"},
			},
		},
		Before: srv.loadConfig,
		Action: srv.runServer(srv),
	}

	if err := app.Run(os.Args); err != nil {
		log.Println(err)
	}
}

func (s *Server) loadConfig(c *cli.Context) error {
	serverURL := c.String("server-url")
	parsedServerURL, err := url.Parse(serverURL)
	if err != nil {
		return err
	}

	s.serverURL = parsedServerURL

	return nil
}

func (s *Server) runServer(srv *Server) cli.ActionFunc {
	return func(c *cli.Context) error {
		srv.sshConfig = &ssh.ServerConfig{
			ServerVersion:     "SSH-2.0-OneAuth (+https://oneauth.vitalvas.dev)",
			PasswordCallback:  srv.sshPasswordCallback,
			PublicKeyCallback: srv.sshPublicKeyCallback,

			BannerCallback: func(conn ssh.ConnMetadata) string {
				remote, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
				return fmt.Sprintf("Welcome %s from %s!\n", conn.User(), remote)
			},
		}

		private, err := generatePrivateHostKey()
		if err != nil {
			log.Fatal(err)
		}

		srv.sshConfig.AddHostKey(private)

		return srv.ListenAndServe()
	}
}

func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", ":2022")
	if err != nil {
		return err
	}

	defer listener.Close()

	log.Printf("listening on %s", ":2022")
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Println("failed to accept connection: ", err)
			continue
		}

		go s.handleConn(tcpConn)
	}
}
