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

const defaultListenAddr = ":2022"

type Server struct {
	serverURL  *url.URL
	sshConfig  *ssh.ServerConfig
	listenAddr string
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
	return func(_ *cli.Context) error {
		if err := srv.setupSSHConfig(); err != nil {
			return err
		}

		return srv.ListenAndServe()
	}
}

func (s *Server) setupSSHConfig() error {
	s.sshConfig = &ssh.ServerConfig{
		ServerVersion:     "SSH-2.0-OneAuth (+https://oneauth.vitalvas.dev)",
		PasswordCallback:  s.sshPasswordCallback,
		PublicKeyCallback: s.sshPublicKeyCallback,

		BannerCallback: func(conn ssh.ConnMetadata) string {
			remote, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
			return fmt.Sprintf("Welcome %s from %s!\n", conn.User(), remote)
		},
	}

	private, err := generatePrivateHostKey()
	if err != nil {
		return err
	}

	s.sshConfig.AddHostKey(private)

	return nil
}

func (s *Server) ListenAndServe() error {
	addr := s.listenAddr
	if addr == "" {
		addr = defaultListenAddr
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	defer listener.Close()

	log.Printf("listening on %s", addr)

	return s.serve(listener)
}

func (s *Server) serve(listener net.Listener) error {
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConn(tcpConn)
	}
}
