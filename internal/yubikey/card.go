package yubikey

import (
	"fmt"
	"strings"

	"github.com/go-piv/piv-go/v2/piv"
)

type Card struct {
	Name    string
	Serial  uint32
	Version string
}

func (c *Card) String() string {
	return fmt.Sprintf("Yubikey #%d", c.Serial)
}

func Cards() ([]Card, error) {
	cards, err := piv.Cards()
	if err != nil {
		return nil, fmt.Errorf("failed to list cards: %w", err)
	}

	response := make([]Card, 0, len(cards))

	for _, name := range cards {
		if !strings.Contains(strings.ToLower(name), "yubikey") {
			continue
		}

		card, err := cardRead(name)
		if err != nil {
			return nil, fmt.Errorf("failed to read card: %w", err)
		}

		response = append(response, *card)
	}

	return response, nil
}

func cardRead(name string) (*Card, error) {
	yk, err := piv.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open card: %w", err)
	}
	defer yk.Close()

	serial, err := yk.Serial()
	if err != nil {
		return nil, fmt.Errorf("failed to get serial: %w", err)
	}

	version := yk.Version()
	versionStr := fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)

	return &Card{
		Name:    name,
		Serial:  serial,
		Version: versionStr,
	}, nil
}
