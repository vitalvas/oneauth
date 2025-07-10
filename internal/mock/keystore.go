package mock

import (
	"github.com/vitalvas/oneauth/internal/keystore"
)

// NewKeystore creates a new keystore instance for testing
func NewKeystore() *keystore.Store {
	return keystore.New(300) // 5 minute keep time
}
