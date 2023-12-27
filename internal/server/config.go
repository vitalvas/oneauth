package server

import (
	"encoding/json"
	"os"

	"github.com/urfave/cli/v2"
)

type Config struct {
	Yubico ConfigYubico `json:"yubico"`
}

type ConfigYubico struct {
	ClientID     int    `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (s *Server) loadConfig(c *cli.Context) error {
	configFile := c.String("config")

	jsonFile, err := os.Open(configFile)
	if err != nil {
		return err
	}

	defer jsonFile.Close()

	return json.NewDecoder(jsonFile).Decode(&s.config)
}
