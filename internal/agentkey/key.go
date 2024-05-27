package agentkey

import (
	"crypto/rand"
	"fmt"

	"github.com/vitalvas/oneauth/internal/tools"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Key struct {
	name        string
	fingerprint string
	signer      ssh.Signer
	agentKey    *agent.Key
}

func NewKey(key agent.AddedKey) (*Key, error) {
	signer, err := ssh.NewSignerFromKey(key.PrivateKey)
	if err != nil {
		return nil, err
	}

	pubKey := signer.PublicKey()
	fingerprint := tools.SSHFingerprint(pubKey)

	var keyName string
	if key.Comment != "" {
		keyName = key.Comment
	} else {
		keyName = fingerprint
	}

	return &Key{
		name:        keyName,
		fingerprint: fingerprint,
		signer:      signer,
		agentKey: &agent.Key{
			Format:  pubKey.Type(),
			Blob:    pubKey.Marshal(),
			Comment: key.Comment,
		},
	}, nil
}

func (k *Key) Fingerprint() string {
	return k.fingerprint
}

func (k *Key) AgentKey() *agent.Key {
	return k.agentKey
}

func (k *Key) Sign(data []byte, flags agent.SignatureFlags) (*ssh.Signature, error) {
	if flags == 0 {
		return k.signer.Sign(rand.Reader, data)
	}

	algorithmSigner, ok := k.signer.(ssh.AlgorithmSigner)
	if !ok {
		return nil, fmt.Errorf("signature does not support non-default signature algorithm: %T", k.signer)
	}

	var algorithm string
	switch flags {
	case agent.SignatureFlagRsaSha256:
		algorithm = ssh.KeyAlgoRSASHA256
	case agent.SignatureFlagRsaSha512:
		algorithm = ssh.KeyAlgoRSASHA512
	default:
		return nil, fmt.Errorf("unsupported signature flags: %d", flags)
	}

	return algorithmSigner.SignWithAlgorithm(rand.Reader, data, algorithm)
}
