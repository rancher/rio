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
	h := md5.Sum([]byte(s))
	d := hex.EncodeToString(h[:])
	return fmt.Sprintf("%s-%s", s[:count-6], d[:5])
}
