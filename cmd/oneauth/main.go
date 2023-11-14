package main

import (
	"log"
	"runtime"

	"github.com/vitalvas/oneauth/internal/commands"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	runtime.GOMAXPROCS(1)
	commands.Execute()
}
