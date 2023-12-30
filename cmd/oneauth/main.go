package main

import (
	"log"
	"runtime"

	"github.com/vitalvas/oneauth/cmd/oneauth/commands"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	runtime.GOMAXPROCS(1)
	commands.Execute()
}
