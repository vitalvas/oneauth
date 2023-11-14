package certgen

import "fmt"

func GenCommonName(name string) string {
	return fmt.Sprintf("oneauth@%s", name)
}
