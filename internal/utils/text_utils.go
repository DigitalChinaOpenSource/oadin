package utils

func SplitIntoChunks(text string, chunkSize int) []string {
	if chunkSize <= 0 {
		return []string{text}
	}

	var chunks []string
	runes := []rune(text)

	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}
