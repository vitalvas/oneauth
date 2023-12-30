package commands

import (
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/yubikey"
)

const infoTmpl = `
--- Keys ---
{{- range $key := .Keys }}
- {{ $key.Name }} (Serial: {{ $key.Serial }}, Version: {{ $key.Version }})
{{- end }}
`

type InfoKey struct {
	Name    string
	Serial  string
	Version string
}

type infoData struct {
	Keys []InfoKey
}

var infoCmd = &cli.Command{
	Name:  "info",
	Usage: "Prints detailed information",
	Action: func(c *cli.Context) error {
		cards, err := yubikey.Cards()
		if err != nil {
			return err
		}

		info := infoData{}

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
