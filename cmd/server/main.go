package main

import (
	"log"

	"github.com/vitalvas/oneauth/internal/server"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	server.Execute()
}
