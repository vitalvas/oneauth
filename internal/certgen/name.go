package certgen

import "fmt"

func GenCommonName(user, name string) string {
	return fmt.Sprintf("%s@%s", user, name)
}
