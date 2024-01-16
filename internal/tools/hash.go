package tools

import (
	"fmt"
	"hash/fnv"
)

func FastHash(data []byte) (string, error) {
	h := fnv.New64a()

	if _, err := h.Write(data); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum64()), nil
}
