package tools

func CountUniqueChars(s string) int {
	charMap := make(map[rune]bool)

	for _, char := range s {
		charMap[char] = true
	}

	return len(charMap)
}
