package stringers

import "strings"

func IsRepo(image string) bool {
	return strings.HasSuffix(image, ".git") ||
		strings.HasPrefix(image, "http:") ||
		strings.HasPrefix(image, "https:")
}
