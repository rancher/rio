package name

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

func Limit(s string, count int) string {
	if len(s) < count {
		return s
	}
	return fmt.Sprintf("%s-%s", s[:count-6], Hex(s, 5))
}

func Hex(s string, length int) string {
	h := md5.Sum([]byte(s))
	d := hex.EncodeToString(h[:])
	return d[:length]
}
