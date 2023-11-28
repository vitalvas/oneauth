package tools

import (
	"log"
	"os"
)

func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	return home
}
