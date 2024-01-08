package keyring

import "fmt"

func GetYubikeyAccount(keyID uint64, name string) string {
	return fmt.Sprintf("yubikey:%d:%s", keyID, name)
}
