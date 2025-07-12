package commands

import (
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/cmd/oneauth/config"
	"github.com/vitalvas/oneauth/cmd/oneauth/rpclient"
	"github.com/vitalvas/oneauth/internal/yubikey"
)

const infoTmpl = `
--- Keys ---
{{- range $key := .Keys }}
- {{ $key.Name }} (Serial: {{ $key.Serial }}, Version: {{ $key.Version }})
{{- end }}
{{- if .AgentPid }}

--- Agent ---
Agent PID: {{ .AgentPid }}
Version: {{ .AgentVersion }}
Uptime: {{ .AgentUptime }}
{{- end }}
`

type InfoKey struct {
	Name    string
	Serial  string
	Version string
}

type infoData struct {
	Keys         []InfoKey
	AgentPid     int
	AgentVersion string
	AgentUptime  string
}

var infoCmd = &cli.Command{
	Name:  "info",
	Usage: "Prints detailed information",
	Before: loadConfig,
	Action: func(c *cli.Context) error {
		info := infoData{}

		// Try to get info from running agent via control socket first
		if globalConfig != nil {
			client, err := rpclient.New(globalConfig.ControlSocketPath)
			if err == nil {
				defer client.Close()
				
				rpcInfo, err := client.GetInfo()
				if err == nil {
					info.AgentPid = rpcInfo.Pid
					info.AgentVersion = rpcInfo.Version
					info.AgentUptime = rpcInfo.Uptime
					// Use keys from the running agent
					for _, key := range rpcInfo.Keys {
						info.Keys = append(info.Keys, InfoKey{
							Name:    key.Name,
							Serial:  key.Serial,
							Version: key.Version,
						})
					}
					
					render := func(tmpl string, data interface{}) string {
						var out strings.Builder
						if err := template.Must(template.New("tmpl").Parse(tmpl)).Execute(&out, data); err != nil {
							log.Fatalf("failed to render template: %v", err)
						}
						return strings.TrimSpace(out.String())
					}

					fmt.Println(render(infoTmpl, info))
					return nil
				}
			}
		}

		// Fallback to local YubiKey detection when control socket is unavailable
		var configPath string
		if globalConfig == nil {
			configPath = c.String("config")
			cfg, err := config.Load(configPath)
			if err != nil {
				// Continue with YubiKey detection even if config fails
				log.Printf("Warning: failed to load config: %v", err)
			} else {
				globalConfig = cfg
			}
		}

		cards, err := yubikey.Cards()
		if err != nil {
			return err
		}

		for _, card := range cards {
			info.Keys = append(info.Keys, InfoKey{
				Name:    card.String(),
				Serial:  fmt.Sprintf("%d", card.Serial),
				Version: card.Version,
			})
		}

		render := func(tmpl string, data interface{}) string {
			var out strings.Builder
			if err := template.Must(template.New("tmpl").Parse(tmpl)).Execute(&out, data); err != nil {
				log.Fatalf("failed to render template: %v", err)
			}
			return strings.TrimSpace(out.String())
		}

		fmt.Println(render(infoTmpl, info))
		return nil
	},
}
