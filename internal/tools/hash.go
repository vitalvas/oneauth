package tools

import (
	"fmt"
	"hash/fnv"
)

func FastHash(data []byte) string {
	h := fnv.New64a()

	h.Write(data)

	return fmt.Sprintf("%x", h.Sum64())
}
