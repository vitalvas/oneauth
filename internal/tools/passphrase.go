package tools

import "crypto/sha256"

func EncodePassphrase(passphrase []byte) []byte {
	hash := sha256.New()
	hash.Write(passphrase)
	return hash.Sum(nil)
}
