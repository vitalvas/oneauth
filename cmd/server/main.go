package main

import (
	"log"

	"github.com/vitalvas/oneauth/cmd/server/server"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	server.Execute()
}
