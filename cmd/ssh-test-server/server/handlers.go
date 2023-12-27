package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

func (s *Server) handleConn(tcpConn net.Conn) {
	defer tcpConn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, s.sshConfig)
	if err != nil {
		if err != io.EOF {
			log.Println("failed to handshake: ", err)
		}
		return
	}

	log.Printf("new connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

	go s.handleRequests(reqs)
	go s.handleChannels(chans)
}

func (s *Server) handleRequests(reqs <-chan *ssh.Request) {
	for req := range reqs {
		log.Printf("request: %v", req)
	}
}

func (s *Server) handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		switch newChannel.ChannelType() {
		case "session":
			s.handleChannelSession(newChannel)

		default:
			newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", newChannel.ChannelType()))
		}
	}
}

func (s *Server) handleChannelSession(newChannel ssh.NewChannel) error {
	conn, requests, err := newChannel.Accept()
	if err != nil {
		return fmt.Errorf("could not accept channel: %v", err)
	}

	defer conn.Close()

	handleShell := func(req *ssh.Request) error {
		defer func() { _ = conn.Close() }()

		if err := req.Reply(true, nil); err != nil {
			return err
		}

		if _, err := conn.Write([]byte("good human!\r\n\r\n")); err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}

		return nil
	}

	for req := range requests {
		switch req.Type {
		case "shell", "exec":
			if err := handleShell(req); err != nil {
				log.Println("failed to handle shell request: ", err)
				if err := req.Reply(false, nil); err != nil {
					return err
				}
			}
		}

		if err := req.Reply(true, nil); err != nil {
			return err
		}
	}

	return nil
}

func generatePrivateHostKey() (ssh.Signer, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	x509Encoded, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})

	return ssh.ParsePrivateKey(pemEncoded)
}
