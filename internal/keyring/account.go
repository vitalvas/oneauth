package keyring

import "fmt"

func GetYubikeyAccount(keyID uint32, name string) string {
	return fmt.Sprintf("yubikey:%d:%s", keyID, name)
}
