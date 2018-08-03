package funcs

import (
	"bytes"
	"unicode"
)

func SplitPreserveQuotes(s string) []string {
	var pieces []string
	var buffer bytes.Buffer
	inQuotes := false

	for i, r := range s {
		if unicode.In(r, unicode.Quotation_Mark) {
			if inQuotes {
				inQuotes = false
			} else {
				inQuotes = true
			}
		}

		lastCharacter := i == len(s)-1
		if (unicode.IsSpace(r) && !inQuotes) || lastCharacter {
			if lastCharacter {
				buffer.WriteRune(r)
			}
			pieces = append(pieces, buffer.String())
			buffer = bytes.Buffer{}
		} else {
			buffer.WriteRune(r)
		}
	}

	return pieces
}
