package server

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	sshConfig *ssh.ServerConfig
}

func Execute() {
	srv := &Server{}

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

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
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
